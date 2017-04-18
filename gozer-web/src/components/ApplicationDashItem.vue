<template>
  <div class="panel panel-default application-card">
    <div class="panel-heading">
      <h3 class="panel-title">
        <router-link :to="'/application/' + applicationName">
          {{applicationName}}
        </router-link>
        <span class="label label-default pull-right">{{applicationMaxAge}}</span>
      </h3>
    </div>
    <div class="panel-body">
      <div v-if="oldestImage">
        <a class="btn btn-default btn-github" target="_blank"
           :href="oldestImage.metadata['git-hub-base'] + '/tree/' + oldestImage.metadata['git-commit']"><img
          src="../img/GitHub-Mark-32px.png" alt="github"/></a>
      </div>
      <div v-else class="text-center middle">
        <p>No metadata for deployment</p>
      </div>
      <spinner v-model="loading" size="sm"></spinner>
    </div>
  </div>
</template>

<script>
  import {accordion, panel, spinner} from 'vue-strap'
  import moment from 'moment';
  import Promise from 'bluebird';
  import imagesService from '../services/ImagesService'
  import _ from 'underscore'

  function cleanUpGitLink(gitLink) {
    return gitLink
      .replace('git@github.com:', 'https://github.com/')
      .replace('.git', '')
  }

  export default {
    props: {
      applicationName: {type: String},
      deployments: {type: Array}
    },
    data() {
      return {
        applicationMaxAge: undefined,
        deploymentsInternal: JSON.parse(JSON.stringify(this.deployments || [])),
        oldestImage: undefined,
        loading: true
      };
    },
    created: function () {
      const vm = this;
      vm.loading = true;
      let containers = _.flatten(vm.deploymentsInternal.map(function (deployment) {
        return deployment.containers
      }))

      let images = []
      return Promise.map(containers, function (container) {
        let containerImageParts = container.image.split(':');
        let containerName = containerImageParts[0];
        let containerVersion = containerImageParts[1];
        return imagesService.getByVersion(containerName, containerVersion)
          .then(function (image) {
            if (image) {

              image.metadata['git-hub-base'] = cleanUpGitLink(image.metadata['git-origin']);
              image.metadata["created-at"] = moment(image.metadata["created-at"]);
              images.push(image)
            }
          })
          .catch(function (error) {
            if (error && error.status !== 404) {
              console.error(error)
            }
          })

      })
        .then(function () {
          if (images.length > 0) {
            let oldest = _.chain(images)
              .sortBy('metadata["created-at"]')
              .first()
              .value()
            vm.applicationMaxAge = oldest.metadata['created-at'].fromNow(); //counter intuitively, dates are smaller when they are older, so we use min
            vm.oldestImage = oldest;
          }
        })
        .finally(function () {
          vm.loading = false;
        });
    },
    components: {
      accordion,
      panel,
      spinner
    }
  };
</script>

<style scoped>

  .btn-github {
    padding: 2px 6px;
    margin: 5px;
  }

  .btn-github img {
    height: 28px;
    width: 28px;
  }

  .label-github a {
    color: white;
  }

  .application-card {
    height: 100%;
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 133px;
    overflow: hidden;
    text-overflow: ellipsis;
    flex-basis: 320px;
    min-width: 320px;

    position: relative;
  }

  .panel-heading {
    flex: 1;
  }

  .panel-body {
    flex: 2;
  }

</style>
