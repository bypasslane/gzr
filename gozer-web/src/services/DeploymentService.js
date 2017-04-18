import _ from 'underscore';
import Vue from 'vue';
import Promise from 'bluebird';
import VueResource from 'vue-resource';
import moment from 'moment';
import imagesService from './ImagesService';
import imageUtilities from '../utilities/ImageUtilities';

Vue.use(VueResource);
const createdAtKey = 'created-at';
const deploymentsResource = Vue.resource('/deployments{/name}');

function transformDeploymentToViewObj(deployment) {
  return {
    $$originalObject: deployment,
    name: deployment.metadata.name,
    uid: deployment.metadata.uid,
    labels: deployment.metadata.labels,
    containers: deployment.spec.template.spec.containers
      .map((container) => {
        return {
          name: container.name,
          image: container.image,
          imageName: container.image.split(':')[0],
          ports: container.ports
        }
      }),
    maxUnresponsiveTimeBeforeTermination: deployment.spec.template.spec.terminationGracePeriodSeconds,
    status: _.omit(deployment.status, 'conditions')
  };
}

function list() {
  return deploymentsResource.get().then(function (response) {
    return response.data.deployments.map(transformDeploymentToViewObj)
  });
}

function get(name) {
  return deploymentsResource.get({name: name}).then(function (response) {
    return transformDeploymentToViewObj(response.data);
  })
}

function getApplicationDeployments(applicationName) {
  return list()
    .then(function (list) {
      //groupBy returns a hash, that is turned into objects in an array, which is then sorted
      let grouped = _.chain(list)
        .groupBy(function (item) {
          return item.labels.application || "Missing Application Label"
        })
        .value();
      return grouped[applicationName];
    })
}

function deploymentsToApplicationDetails(deployments) {
  let images = getUniqueImagesForDeployment(deployments)
  let imageLists = {};

  return Promise.map(images,
    function (imageName) {
      return imagesService
        .get(imageName)
        .then(function (list) {
          imageLists[imageName] = imageUtilities.sortImagesByCreatedAtDateDesc(list);
        })
        .catch(function () {
          //adding in a blank array that provides at least enough info for displaying what we do know without metadata
          imageLists[imageName] = [];
        })
    })
    .then(function () {
      //sorts the deployments as they are added to the object async, which can cause them to be added to the object out of order.
      return sortObjectProperties(imageLists);
    })
    .then(function (imageLists) {
      associateDeploymentsToImages(deployments, imageLists);
      return {deployments: deployments, imageLists: imageLists};
    })
}

function getUniqueImagesForDeployment(deployments) {
  return _.chain(deployments)
    .map(function (deployment) {
      return deployment.containers.map(function (container) {
        return container.imageName;
      })
    })
    .flatten()
    .uniq()
    .value()
}

function sortObjectProperties(object) {
  let result = {}
  _.chain(object)
    .keys()
    .sortBy()
    .each(function (key) {
      result[key] = object[key];
    })
    .value();
  return result;
}

function associateDeploymentsToImages(deployments, imageLists) {
  deployments.forEach(function (deployment) {
    deployment.containers.forEach(function (container) {
      let imageToFind = container.image;
      let images = imageLists[container.imageName];
      let foundImage = _.findWhere(images, {name: imageToFind})

      if (foundImage) {
        foundImage.deployments = foundImage.deployments || [];
        foundImage.deployments.push(deployment);
      }
      else {
        //adding in a blank object that provides at least enough info for displaying what we do know without metadata
        images.push({
          name: imageToFind,
          metadata: {
            'created-at': moment.invalid()
          },
          deployments: [deployment]
        })
      }
    })
  })
}

function getApplicationDetails(applicationName) {
  return getApplicationDeployments(applicationName)
    .then(deploymentsToApplicationDetails);
}

function getDeploymentWithImageData(name) {
  return get(name)
    .then(function (deployment) {
      return Promise.reduce(
        deployment.containers,
        function (image, container) {
          if (image) {
            return image;
          }
          else {
            let containerImageParts = container.image.split(':');
            let containerName = containerImageParts[0];
            let containerVersion = containerImageParts[1];

            return imagesService
              .getByVersion(containerName, containerVersion)
              .then(function (imageData) {
                imageData.containerName = container.name;
                return imageData;
              })
          }
        }, null).then(function (image) {
        if (image) {
          return imagesService
            .get(image.containerName)
            .then(function (list) {
              list.forEach(function (image) {
                let imageParts = image.name.split(':');
                let imageName = imageParts[0];
                let imageVersion = imageParts[1];
                image.age = image.metadata["created-at"].fromNow();
                image.deploymentImageName = [imageName, imageVersion].join(':');
              });
              list.sort(function (leftImage, rightImage) {
                return leftImage.metadata[createdAtKey].isBefore(rightImage.metadata[createdAtKey])
              });
              return {
                deployment: deployment,
                deploymentAppImage: image,
                deploymentAppImageName: image.name,
                deploymentImages: list
              }
            });
        }
        else {
          Promise.reject('No image data found');
        }
      });
    });
}

function set(deploymentName, container_name, image_name) {
  return deploymentsResource
    .update({name: deploymentName},
      {
        container_name: container_name,
        image: image_name
      })
}
function setImageForAllDeployments(deployments, imageObject, imageName) {
  return Promise.map(deployments, function (deployment) {
    let containerName = _.chain(deployment.containers)
        .findWhere({imageName: imageName})
        .value()
        .name || undefined;
    if (containerName) {
      return set(deployment.name, containerName, imageObject.name)
    }
    return Promise.resolve()
  })
}

export default {
  list,
  get,
  set,
  setImageForAllDeployments,
  getDeploymentWithImageData,
  getApplicationDeployments,
  getApplicationDetails
}

