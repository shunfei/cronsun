window.$ = window.jQuery = require('jquery');
require('semantic');
require('semantic-ui/dist/semantic.min.css');

import Vue from 'vue';
Vue.config.debug = true;

// global restful client
import Rest from './libraries/rest-client.js';
const RestApi =(Vue, options)=>{
  Vue.prototype.$rest = new Rest('/v1/');
};
Vue.use(RestApi);

// global event bus
Vue.use((Vue)=>{
  Vue.prototype.$bus = new Vue();
});

import VueRouter from 'vue-router';
Vue.use(VueRouter);

import App from './App.vue';
import Dash from './components/Dash.vue';
import Log from './components/Log.vue';
import LogDetail from './components/LogDetail.vue';
import Job from './components/Job.vue';
import JobEdit from './components/JobEdit.vue';
import Node from './components/Node.vue';

var routes = [
  {path: '/', component: Dash},
  {path: '/log', component: Log},
  {path: '/log/:id', component: LogDetail},
  {path: '/job', component: Job},
  {path: '/job/create', component: JobEdit},
  {path: '/job/edit/:group/:id', component: JobEdit},
  {path: '/node', component: Node}
];

var router = new VueRouter({
  routes: routes
});

var app = new Vue({
  el: '#app',
  render: h => h(App),
  router: router
});