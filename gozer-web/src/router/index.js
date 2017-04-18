import Vue from 'vue';
import Router from 'vue-router';
import Index from 'components/Index';
import Deployment from 'components/Deployment';
import Application from 'components/Application';

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index
    },
    {
      path:'/application/:name',
      name:'Application',
      component: Application
    },
    {
      path:'/deployment/:name',
      name:'Deployment',
      component: Deployment
    }
  ],
});
