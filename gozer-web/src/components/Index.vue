<template>
  <div>
    <div v-if="loading">
      <spinner v-model="loading" size="lg" fixed text="Loading Deployments"></spinner>
    </div>
    <div v-else-if="!error">
      <div class="gzr-cards">
        <div class="flex-row">
          <div v-for="group in groupedList">
            <application-dash-item :deployments="group.deployments" :application-name="group.appName"></application-dash-item>
          </div>
        </div>
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
              Sorry there was an error loading deployments.
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
  import {spinner} from 'vue-strap'
  import DeploymentService from '../services/DeploymentService'
  import ApplicationDashItem from './ApplicationDashItem'
  import _ from 'underscore'

  export default {
    name: 'index',
    data ()  {
      return {
        list: [],
        loading: true,
        error: null
      }
    },
    created() {
      const vm = this;
      DeploymentService
        .list()
        .then(function (list) {
          //groupBy returns a hash, that is turned into objects in an array, which is then sorted
          let grouped = _.chain(list)
            .groupBy(function (item) {
              return item.labels.application || "Missing Application Label"
            })
            .mapObject(function (val, key) {
              return {deployments: val, appName: key}
            })
            .sortBy(function (item) {
              return item.appName.toLocaleLowerCase();
            })
            .value();

          vm.groupedList = grouped;
        })
        .catch(function (err) {
          vm.error = err;
        })
        .finally(function () {
          vm.loading = false;
        })
    },
    components: {
      ApplicationDashItem,
      spinner
    }
  };
</script>
<style scoped>
  .gzr-cards {
    display: flex;
    flex-direction: column;
  }

  .flex-row {
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: space-around;

  }

  .flex-row > div {
    margin-bottom: 15px;
  }

</style>
