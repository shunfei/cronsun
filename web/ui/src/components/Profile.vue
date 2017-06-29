<template>
  <form v-else class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <div class="field">
      <label>{{$L('password')}}</label>
      <input type="password" v-model:value="password"/>
    </div>
    <div class="field">
      <label>{{$L('new password')}}</label>
      <input type="password" v-model:value="newPassword"/>
    </div>
    <div class="field">
      <button class="fluid blue ui button" type="button" v-on:click="submit">{{$L('save')}}</button>
    </div>
  </form>
</template>

<script>
export default {
  name: 'profile',

  data: function(){
    return {
      loading: false,
      password: '',
      newPassword: ''
    }
  },

  methods: {
    submit(){
      var vm = this;
      this.loading = true;
      
      this.$rest.POST('user/setpwd', {password: this.password, newPassword: this.newPassword}).
        onsucceed(200, (resp)=>{
          vm.$bus.$emit('success', vm.$L('your password has been change'));
        }).
        onfailed((msg)=>{vm.$bus.$emit('error', msg)}).
        onend(()=>{vm.loading = false}).
        do();
    }
  }
}
</script>
