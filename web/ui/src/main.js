window.$ = window.jQuery = require('jquery');
require('semantic');
require('semantic-ui/dist/semantic.min.css');

import Vue from 'vue';
Vue.config.debug = true;

import Lang from './i18n/language';
Vue.use((Vue)=>{
  Vue.prototype.$L = Lang.L
  Vue.prototype.$Lang = Lang
});

// global restful client
import Rest from './libraries/rest-client.js';
var restApi = new Rest('/v1/');
const RestApi = (Vue, options)=>{
  Vue.prototype.$rest = restApi;
};
Vue.use(RestApi);

// global event bus
var bus = new Vue();
Vue.use((Vue)=>{
  Vue.prototype.$bus = bus;
});

import VueRouter from 'vue-router';
Vue.use(VueRouter);

import App from './App.vue';
import Dash from './components/Dash.vue';
import Log from './components/Log.vue';
import LogDetail from './components/LogDetail.vue';
import Job from './components/Job.vue';
import JobEdit from './components/JobEdit.vue';
import JobExecuting from './components/JobExecuting.vue';
import Node from './components/Node.vue';
import NodeGroup from './components/NodeGroup.vue';
import NodeGroupEdit from './components/NodeGroupEdit.vue';

var routes = [
  {path: '/', component: Dash},
  {path: '/log', component: Log},
  {path: '/log/:id', component: LogDetail},
  {path: '/job', component: Job},
  {path: '/job/create', component: JobEdit},
  {path: '/job/edit/:group/:id', component: JobEdit},
  {path: '/job/executing', component: JobExecuting},
  {path: '/node', component: Node},
  {path: '/node/group', component: NodeGroup},
  {path: '/node/group/create', component: NodeGroupEdit},
  {path: '/node/group/:id', component: NodeGroupEdit}
];

var router = new VueRouter({
  routes: routes
});


restApi.GET('configurations').onsucceed(200, (resp)=>{
  const Config = (Vue, options)=>{
    Vue.prototype.$appConfig = resp;
  }
  Vue.use(Config);

  restApi.defaultExceptionHandler = (msg)=>{bus.$emit('error', msg)};
  restApi.defaultFailedHandler = (msg)=>{bus.$emit('error', msg)};
  
  var app = new Vue({
    el: '#app',
    render: h => h(App),
    router: router
  });
}).onfailed((data, xhr)=>{
  var msg = data ? data : xhr.status+' '+xhr.statusText;
  showInitialError('Failed to get global configurations('+xhr.responseURL+'): '+msg);
}).onexception((msg)=>{
  showInitialError('Failed to get global configurations('+xhr.responseURL+'): '+msg);
}).do();

function showInitialError(msg) {
  var d = document.getElementById('app');
  d.innerHTML = msg;
  d.className = 'initial error';
}
