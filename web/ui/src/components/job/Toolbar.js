define(['Vue'], function(Vue){
  Vue.component('job-toolbar', {
    props: [],
    template: '<router-link class="ui button" to="/job/create">新建任务</router-link>'
  });
});