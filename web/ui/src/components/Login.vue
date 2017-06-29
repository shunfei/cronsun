<style scoped>
@media (min-width: 600px) {
  #loginForm {
    width: 500px;
    margin: 100px auto 0;
  }
}
</style>

<template>
  <div>
    <form id="loginForm" class="ui form" v-on:submit.prevent="onSubmit">
      <div class="field">
        <label>{{$L('email')}}</label>
        <input type="text" v-model="email" placeholder:="$L('email')">
      </div>
      <div class="field">
        <label>{{$L('password')}}</label>
        <input type="password" v-model="password" placeholder:="$L('password')">
      </div>
      <div class="field">
        <div class="ui checkbox">
          <input type="checkbox" v-model="remember" tabindex="0" class="hidden">
          <label>{{$L('remember me')}}</label>
        </div>
      </div>
      <button class="ui button" type="submit">{{$L('login')}}</button>
    </form>
  </div>
</template>

<script>
import Vue from 'vue';

export default {
  name: 'login',

  data: function(){
    return {
      email: '',
      password: '',
      remember: false
    }
  },

  methods: {
    onSubmit () {
      var vm = this;

      this.$rest.GET('session?email='+this.email+'&password='+this.password).
      onsucceed(200, (resp)=>{
        vm.$store.commit('setEmail', resp.email);
        vm.$store.commit('setRole', resp.role);
        vm.$store.commit('enabledAuth', resp.enabledAuth);
        vm.$router.push('/')
        vm.getConfig();
      }).
      do();
    },

    getConfig() {
      this.$rest.GET('configurations').onsucceed(200, (resp)=>{
        const Config = (Vue, options)=>{
          Vue.prototype.$appConfig = resp;
        }
        Vue.use(Config);
      }).onfailed((data, xhr)=>{
        var msg = data ? data : xhr.status+' '+xhr.statusText;
        vm.$bus.$emit('error', msg);
      }).do();
    }
  }
}
</script>
