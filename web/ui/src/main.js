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

// global event bus
var bus = new Vue();
Vue.use((Vue)=>{
  Vue.prototype.$bus = bus;
});

// global restful client
import Rest from './libraries/rest-client.js';
var restApi = new Rest('/v1/', (msg) => {
  bus.$emit('error', msg);
}, (msg)=>{
  bus.$emit('error', msg);
}, {
  401: (data, xhr) => {bus.$emit('goLogin')}
});
Vue.use((Vue, options)=>{
  Vue.prototype.$rest = restApi;
}, null);

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
import Account from './components/Account.vue';
import AccountEdit from './components/AccountEdit.vue';
import Profile from './components/Profile.vue';
import Login from './components/Login.vue';

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
  {path: '/node/group/:id', component: NodeGroupEdit},
  {path: '/admin/account/list', component: Account},
  {path: '/admin/account/add', component: AccountEdit},
  {path: '/admin/account/edit', component: AccountEdit},
  {path: '/user/setpwd', component: Profile},
  {path: '/login', component: Login}
];

var router = new VueRouter({
  routes: routes
});

var app = new Vue({
  el: '#app',
  render: h => h(App),
  router: router
});
