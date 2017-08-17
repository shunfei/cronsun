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
      <div class="field" ref="remember">
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
      remember: ''
    }
  },

  mounted: function(){
    var vm = this;
    $(this.$refs.remember).find('.checkbox').checkbox({
      onChange: function(){
        vm.remember = $(vm.$refs.remember).find('input[type=checkbox]:checked').val();
      }
    });
  },

  methods: {
    onSubmit(){
      var vm = this;

      this.$rest.GET('session?email='+this.email+'&password='+this.password+'&remember='+this.remember).
      onsucceed(200, (resp)=>{
        vm.$store.commit('setEmail', resp.email);
        vm.$store.commit('setRole', resp.role);
        vm.$store.commit('enabledAuth', resp.enabledAuth);
        vm.$router.push('/')
        vm.$loadConfiguration();
      }).
      do();
    }
  }
}
</script>
