import moment from 'moment';
import _ from 'underscore';

const createdAtKey = 'created-at';
const gitHubBaseKey = 'git-hub-base';
const gitOriginKey = 'git-origin';

function cleanUpGitLink(gitLink) {
  return gitLink
    .replace('git@github.com:', 'https://github.com/')
    .replace('.git', '')
}

function enhanceImage(image){
  image.metadata[gitHubBaseKey] = cleanUpGitLink(image.metadata[gitOriginKey]);
  image.metadata[createdAtKey] = moment(image.metadata[createdAtKey]);
  return image;
}

function sortImagesByCreatedAtDateDesc(imageList) {
  return _.chain(imageList)
    .sortBy(function (image) {
      return image.metadata[createdAtKey];
    })
    .reverse()
    .value();
}

export default {enhanceImage, sortImagesByCreatedAtDateDesc}
