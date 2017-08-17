<template>
  <form v-else class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <h3 class="ui header">{{$L(action == 'CREATE' ? 'add account' : 'edit account')}}</h3>
    <div class="field">
      <label>{{$L('email')}}</label>
      <input type="text" v-model:value="account.email"/>
    </div>
    <div class="field">
      <label>{{$L('password')}}</label>
      <input type="text" v-model:value="account.password"/>
    </div>
    <div class="inline fields" ref="role">
      <label>{{$L('role')}}</label>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" name="role" value="1" v-model="account.role" class="hidden">
          <label>Administrator</label>
        </div>
      </div>
      <div class="field">
        <div class="ui radio checkbox">
          <input type="radio" name="role" value="2" v-model="account.role" class="hidden">
          <label>Developer</label>
        </div>
      </div>
    </div>
    <div v-show="action === 'UPDATE'">
      <div class="inline fields" ref="status">
        <label>{{$L('status')}}</label>
        <div class="field">
          <div class="ui radio checkbox">
            <input type="radio" name="status" value="1" v-model="account.status" class="hidden">
            <label>{{$L('active')}}</label>
          </div>
        </div>
        <div class="field">
          <div class="ui radio checkbox">
            <input type="radio" name="status" value="-1" v-model="account.status" class="hidden">
            <label>{{$L('ban')}}</label>
          </div>
        </div>
      </div>
    </div>
    <div class="field">
      <button class="fluid blue ui button" type="button" v-on:click="submit">{{$L('save')}}</button>
    </div>
  </form>
</template>

<script>
export default {
  name: 'account-edit',

  data: function(){
    return {
      loading: true,
      action: '',
      account: {
        originEmail: '',
        email: '',
        password: '',
        role: 1,
        status: 1
      }
    }
  },

  mounted: function(){
    var vm = this;

    if (this.$route.path.indexOf('/admin/account/add') === 0) {
      this.action = 'CREATE';
      this.loading = false;
    } else {
      this.action = 'UPDATE';
      this.$rest.GET('admin/account/'+this.$route.query.email).
        onsucceed(200, (resp)=>{
          vm.account = {
            originEmail: resp.email,
            email: resp.email,
            password: '',
            role: resp.role,
            status: resp.status,
          }
        }).
        onfailed((msg)=>{vm.$bus.$emit('error', msg)}).
        onend(()=>{vm.loading = false}).
        do();
    }

    $(this.$refs.role).find('.checkbox').checkbox({
      onChange: function(){
        vm.account.role = +$(vm.$refs.role).find('input[type=radio]:checked').val();
      }
    });

    if (this.action === 'UPDATE') {
      $(this.$refs.status).find('.checkbox').checkbox({
        onChange: function(){
          vm.account.status = +$(vm.$refs.status).find('input[type=radio]:checked').val();
        }
      });
    }
  },

  methods: {
    submit(){
      var vm = this, req, code;
      this.loading = true;
      if (this.action === 'CREATE') {
        req = this.$rest.PUT('admin/account', this.account);
        code =  204;
      } else {
        req = this.$rest.POST('admin/account', this.account)
        code =  200;
      }

      req.onsucceed(code, (resp)=>{
          vm.$router.push('/admin/account/list');
        }).
        onfailed((msg)=>{vm.$bus.$emit('error', msg)}).
        onend(()=>{vm.loading = false}).
        do();
    }
  }
}
</script>
