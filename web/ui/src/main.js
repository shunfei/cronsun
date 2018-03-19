window.$ = window.jQuery = require('jquery');
require('semantic');
require('semantic-ui/dist/semantic.min.css');
import store from './vuex/store';

import Vue from 'vue';
Vue.config.debug = true;

import Lang from './i18n/language';
Vue.use((Vue) => {
  Vue.prototype.$L = Lang.L
  Vue.prototype.$Lang = Lang
});

// global event bus
var bus = new Vue();
Vue.use((Vue) => {
  Vue.prototype.$bus = bus;
});

// global restful client
import Rest from './libraries/rest-client.js';
var restApi = new Rest('/v1/', (msg) => {
  bus.$emit('error', msg);
}, (msg) => {
  bus.$emit('error', msg);
}, {
    401: (data, xhr) => { bus.$emit('goLogin') }
  });
Vue.use((Vue, options) => {
  Vue.prototype.$rest = restApi;
}, null);

import VueRouter from 'vue-router';
Vue.use(VueRouter);

Vue.use((Vue) => {
  Vue.prototype.$loadConfiguration = () => {
    restApi.GET('configurations').
      onsucceed(200, (resp) => {
        const Config = (Vue, options) => {
          Vue.prototype.$appConfig = resp;
        }
        Vue.use(Config);
        bus.$emit('conf_loaded', resp);
      }).onfailed((data, xhr) => {
        var msg = data ? data : xhr.status + ' ' + xhr.statusText;
        bus.$emit('error', msg);
      }).do();
  }
});

const onConfigLoaded = (Vue, options) => {
  let loaded = false;
  let queue = [];
  let appConfig;

  Vue.prototype.$onConfigLoaded = (f) => {
    if (loaded) {
      f(appConfig);
      return;
    }
    queue.push(f);
  }

  bus.$on('conf_loaded', (c) => {
    loaded = true;
    appConfig = c;
    queue.forEach((f) => {
      f(appConfig)
    })
  });
}
Vue.use(onConfigLoaded);

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
  { path: '/', component: Dash },
  { path: '/log', component: Log },
  { path: '/log/:id', component: LogDetail },
  { path: '/job', component: Job },
  { path: '/job/create', component: JobEdit },
  { path: '/job/edit/:group/:id', component: JobEdit },
  { path: '/job/executing', component: JobExecuting },
  { path: '/node', component: Node },
  { path: '/node/group', component: NodeGroup },
  { path: '/node/group/create', component: NodeGroupEdit },
  { path: '/node/group/:id', component: NodeGroupEdit },
  { path: '/admin/account/list', component: Account },
  { path: '/admin/account/add', component: AccountEdit },
  { path: '/admin/account/edit', component: AccountEdit },
  { path: '/user/setpwd', component: Profile },
  { path: '/login', component: Login }
];

var router = new VueRouter({
  routes: routes
});

bus.$on('goLogin', () => {
  store.commit('setEmail', '');
  store.commit('setRole', 0);
  router.push('/login');
});

var initConf = new Promise((resolve) => {
  restApi.GET('session?check=1').
  onsucceed(200, (resp) => {
    store.commit('enabledAuth', resp.enabledAuth);
    store.commit('setEmail', resp.email);
    store.commit('setRole', resp.role);

    restApi.GET('version').onsucceed(200, (resp)=>{
      store.commit('setVersion', resp);
    }).do();
  
    restApi.GET('configurations').
    onsucceed(200, (resp) => {
      Vue.use((Vue) => Vue.prototype.$appConfig = resp);
      bus.$emit('conf_loaded', resp);

      restApi.GET('nodes').onsucceed(200, (resp)=>{
        var nodes = {};
        for (var i in resp) {
          nodes[resp[i].id] = resp[i];
        }
        store.commit('setNodes', nodes);
        resolve();
      }).do();
    }).onfailed((data, xhr) => {
      bus.$emit('error', data ? data : xhr.status + ' ' + xhr.statusText);
      resolve();
    }).do();
  }).onfailed((data, xhr) => {
    if (xhr.status !== 401) {
      bus.$emit('error', data);
    } else {
      store.commit('enabledAuth', true);
    }
    router.push('/login');
    resolve()
  }).
  do();
})

initConf.then(() => {
  new Vue({
    el: '#app',
    render: h => h(App),
    router: router
  });
})



