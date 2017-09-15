<template>
  <div>
    <ol class="breadcrumb">
      <li>
        <router-link to="/">&lt; Back</router-link>
      </li>
      <li v-if="deployment">
        <span class="active">{{deployment.name}}</span>
      </li>
    </ol>
    <div v-if="loading">
      <spinner v-model="loading" size="lg" fixed text="Loading Deployments"></spinner>
    </div>
    <div v-else-if="deploying">
      <spinner v-model="deploying" size="lg" fixed text="Deploying..."></spinner>
    </div>
    <div v-else-if="!error" class="panel panel-default">
      <div class="panel-heading">
        <h2>{{deployment.name}}</h2>
      </div>
      <div class="panel-body">
        <div class="list-group">
          <div class="list-group-item" :class="{active: image.name === deploymentAppImageName}" v-for="image in deploymentImages">
            <h4 class="list-group-item-heading">{{image.name}}<span class="label label-default pull-right">{{image.age}}</span></h4>
            <div class="list-group-item-text">
              <div class="row">
                <div class="col-xs-10">
                  <dl class="dl-horizontal">
                    <dt>Commit</dt>
                    <dd>{{image.metadata['git-commit']}} <a class="btn btn-default btn-github"
                                                            :href="deploymentAppImage.metadata['git-origin'] + '/commit/' +image.metadata['git-commit']"><img
                      src="../img/GitHub-Mark-32px.png" alt="github"/></a></dd>

                    <dt>Annotations</dt>
                    <dd><span class="label label-info" v-for="annotation in image.metadata['git-annotation']">{{annotation}}</span></dd>

                    <dt>Tags</dt>
                    <dd><span class="label label-info" v-for="tag in image.metadata['git-tag']">{{tag}}</span></dd>
                  </dl>
                </div>
                <div class="col-xs-2">
                  <button class="btn btn-primary" v-if="image.name !== deploymentAppImageName" @click="doDeploy(image)">Deploy</button>
                  <span class="live" v-else><span class="glyphicon glyphicon-saved text-success"></span> Live</span>
                </div>
              </div>

            </div>
          </div>
        </div>
        <span></span>
      </div>
    </div>
    <div v-else>
      <div class="row">
        <div class="col-md-4 col-md-offset-4 col-xs-10 col-xs-offset-1">
          <div class="media">
            <div class="media-left media-middle">
              <div class="media-object" style="width: 300px">
                <div style="width:100%;height:0;padding-bottom:55%;position:relative;"><iframe src="http://tv.giphy.com/sad" width="100%" height="100%" style="position:absolute" frameBorder="0" class="giphy-embed"></iframe></div>
              </div>
            </div>
            <div class="media-body media-middle">
              <h3 class="media-heading">Gozer Error</h3>
              Sorry there was an error loading {{name}}.
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
  import { accordion, panel, spinner } from 'vue-strap'
  import Multiselect from 'vue-multiselect'
  import moment from 'moment';
  import imagesService from '../services/ImagesService'
  import deploymentService from '../services/DeploymentService'
  import Promise from 'bluebird';

  export default {
    data() {
      return {
        loading: true,
        error: undefined,
        deploymentAppImage: {},
        deploymentImages: [],
        deploymentAppImageName: '',
        deployment: {},
        deploying: false
      }
    },
    watch: {
      // call the method if the route changes
      '$route': 'fetchData'
    },
    methods: {
      doDeploy: function (image) {
        const vm = this;
        vm.deploying = image;
        deploymentService
          .set(vm.deployment.name, vm.deploymentAppImage.containerName, image.deploymentImageName)
          .then(function () {
            return vm.fetchData();
          })
          .finally(function () {
            vm.deploying = false;
          })
      },
      fetchData: function () {
        const vm = this;
        vm.name =  vm.$route.params.name;
        return Promise.delay(250,
                             deploymentService
                               .getDeploymentWithImageData(vm.name)
                               .then(function (result) {
                                 vm.deployment = result.deployment;
                                 vm.deploymentAppImage = result.deploymentAppImage;
                                 vm.deploymentAppImageName = result.deploymentAppImageName;
                                 vm.deploymentImages = result.deploymentImages;
                               }))
                      .catch(function (error) { if (error && error.status !== 404) { console.log(error) }})
                      .finally(function () {
                        vm.loading = false;
                      });
      }
    },
    created: function () {
      this.fetchData();
    },
    components: {
      accordion,
      panel,
      spinner,
      Multiselect
    }
  };
</script>

<style scoped>

  .live{
    display: inline-block;
    margin-bottom: 0;
    text-align: center;
    vertical-align: middle;
    background-color:#5cb85c;
    color:#fff;
    font-weight: 500;
    border: 1px solid transparent;
    white-space: nowrap;
    padding: 6px 12px;
    font-size: 14px;
    line-height: 1.42857;
    border-radius: 4px;
  }

  .btn-github {
    padding: 0 3px;
    margin: 0 5px;
    display: inline-block;
  }

  .btn-github img {
    height: 16px;
    width: 16px;
    display: inline-block;
  }

  .label-github a {
    color: white;
  }
</style>
