<template>
  <div>
    <div v-if="$store.getters.role === 1">
      <div class="clearfix" style="margin-bottom: 20px;">
        <router-link class="ui right floated primary button" to="/admin/account/add"><i class="add user icon"></i> {{$L('add account')}}</router-link>
      </div>

      <table class="ui hover teal table">
        <thead>
          <tr>
            <th>{{$L('email')}}</th>
            <th>{{$L('role')}}</th>
            <th>{{$L('status')}}</th>
            <th>{{$L('added time')}}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(account, index) in list">
            <td><router-link :to="'/admin/account/edit?email='+account.email">{{account.email}}</router-link></td>
            <td>{{roleMap(account.role)}}</td>
            <td>{{$L(statusMap(account.status))}}</td>
            <td>{{account.createTime}}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else>Access deny.</div>
  </div>
</template>

<script>
export default {
  name: 'account',

  data(){
    return {
      list: []
    }
  },
  
  mounted(){
    var vm = this;
    this.$rest.GET('/admin/accounts').
      onsucceed(200, (resp)=>{
        vm.list = resp;

        vm.$nextTick(()=>{
          $(vm.$el).find('table .ui.dropdown').dropdown();
        });
      }).
      do();
  },

  methods: {
    statusMap(s) {
      switch (s){
        case -1: return 'banned';
        case 1: return 'actived';
      }
      return '';
    },

    roleMap(s) {
      switch (s){
        case 1: return 'Administrator';
        case 2: return 'Developer';
      }
      return '';
    }
  }
}
</script>
