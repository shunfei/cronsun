define(['Vue'], function(Vue){
  Vue.component('sm-button', {
    props: ['name'],
    template: '<button class="ui button">{{name}}</button>'
  });
});