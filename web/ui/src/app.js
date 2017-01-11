require.config({
  baseUrl: './src/components',
  paths: {
    'jquery': '../../vendors/jquery.min',
    'semantic': '../../vendors/semantic/semantic.min',
    'text': '../../vendors/text',
    'Vue': '../..//vendors/vue',
    'VueRouter': '../../vendors/vue-router'
  },
  shim: {
    'semantic': {
      deps: ['jquery'],
      VueRouter: ['Vue']
    }
  }
});

require(['Vue', 'VueRouter', 'dash/Dash', 'log/Log', 'job/Job', 'job/JobEdit', 'node/Node'],
  function(Vue, VueRouter, Dash, Log, Job, JobEdit, Node){
  Vue.use(VueRouter);

  var routes = [
    {path: '/', component: Dash},
    {path: '/log', component: Log},
    {path: '/job', component: Job},
    {path: '/job/create', component: JobEdit},
    {path: '/node', component: Node}
  ];

  var router = new VueRouter({
    routes: routes
  });

  var app = new Vue({
    router: router
  }).$mount('#app');
});