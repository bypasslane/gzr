<template>
  <div>
    <h2>{{applicationName}}</h2>
    <div class="panel panel-default application-card"
         v-for="deploymentImageList, deploymentImageName in applicationDetails.imageLists">
      <div class="panel-heading">
        <h3 class="panel-title">
          Image: {{deploymentImageName}}
        </h3>
      </div>
      <div class="panel-body">
        <div class="list-group">
          <div class="list-group-item"
               v-for="image in deploymentImageList" :class="{active: image.deployments}">
            <h4 class="list-group-item-heading">
              {{image.name}}
              <span class="label label-default pull-right"
                    :title="image.metadata['created-at'].format('LLLL')">{{image.metadata["created-at"].fromNow()}}</span>
            </h4>
            <div class="list-group-item-text">
              <div class="row">
                <div class="col-xs-10">
                  <dl class="dl-horizontal">
                    <dt>Deployments</dt>
                    <dd>
                      <span class="label label-default"
                            v-for="deployment in image.deployments">{{deployment.name}}</span>
                    </dd>
                    <dt>Commit</dt>
                    <dd>{{image.metadata['git-commit']}} <a v-if="image.metadata['git-commit']" class="btn btn-default btn-github" target="_blank"
                                                            :href="image.metadata['git-hub-base'] + '/tree/' + image.metadata['git-commit']"><img
                      src="../img/GitHub-Mark-32px.png" alt="github"/></a></dd>

                    <dt>Annotations</dt>
                    <dd><span class="label label-info" v-for="annotation in image.metadata['git-annotation']">{{annotation}}</span>
                    </dd>

                    <dt>Tags</dt>
                    <dd><span class="label label-info" v-for="tag in image.metadata['git-tag']">{{tag}}</span></dd>
                  </dl>
                </div>
                <div class="col-xs-2">
                  <button class="btn btn-primary"
                          v-if="!(image.deployments && applicationDetails.deployments.length === image.deployments.length)"
                          @click="doDeploy(image, deploymentImageName)">Deploy
                  </button>
                  <span v-else class="live"><span class="glyphicon glyphicon-saved text-success"></span> Live</span>
                </div>
              </div>

            </div>
          </div>
        </div>
        <div v-if="loading">
          <spinner v-model="loading" size="lg"></spinner>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import {accordion, panel, spinner} from 'vue-strap'
  import moment from 'moment';
  import Promise from 'bluebird';
  import DeploymentService from '../services/DeploymentService'
  import imagesService from '../services/ImagesService'
  import _ from 'underscore'

  export default {
    data() {
      return {
        loading: true,
        deploying: false,
        applicationName: '',
        images: {},
        applicationMaxAge: undefined,
        deployments: {},
        applicationDetails: {},
        containers: []
      };
    },
    watch: {
      '$route': 'fetchData'
    },
    methods: {
      doDeploy: function (image, deploymentImageName) {
        const vm = this;
        vm.deploying = true;
        return DeploymentService
          .setImageForAllDeployments(vm.applicationDetails.deployments, image, deploymentImageName)
          .then(function () {
            return vm.fetchData();
          })
          .finally(function () {
            vm.deploying = false;
          })
      },
      fetchData: function () {
        const vm = this;
        vm.loading = true;
        vm.applicationName = vm.$route.params.name;

        return DeploymentService
          .getApplicationDetails(vm.applicationName)
          .then(function (applicationDetails) {
            vm.applicationDetails = applicationDetails;
          })
          .finally(function () {
            vm.loading = false;
          });
      }
    },
    created: function () {
      return this.fetchData();
    },
    components: {
      accordion,
      panel,
      spinner
    }
  };
</script>

<style scoped>

  .live {
    display: inline-block;
    margin-bottom: 0;
    text-align: center;
    vertical-align: middle;
    background-color: #5cb85c;
    color: #fff;
    font-weight: 500;
    border: 1px solid transparent;
    white-space: nowrap;
    padding: 6px 12px;
    font-size: 14px;
    line-height: 1.42857;
    border-radius: 4px;
  }

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
  }

  .panel-heading {
    flex: 1;
  }

  .panel-body {
    flex: 2;
  }

</style>
