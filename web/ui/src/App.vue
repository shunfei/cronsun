<template>
  <div id="app">
    <div class="ui blue inverted menu fixed">
      <div class="item">CRONSUN</div>
      <router-link v-if="shouldOpen" class="item" to="/" v-bind:class="{active: this.$route.path == '/'}"><i class="dashboard icon"></i> {{$L('dashboard')}}</router-link>
      <router-link v-if="shouldOpen" class="item" to="/log" v-bind:class="{active: this.$route.path.indexOf('/log') === 0}"><i class="file text icon"></i> {{$L('log')}}</router-link>
      <router-link v-if="shouldOpen" class="item" to="/job" v-bind:class="{active: this.$route.path.indexOf('/job') === 0}"><i class="calendar icon"></i> {{$L('job')}}</router-link>
      <router-link v-if="shouldOpen" class="item" to="/node" v-bind:class="{active: this.$route.path.indexOf('/node') === 0}"><i class="server icon"></i> {{$L('node')}}</router-link>
      <router-link v-if="$store.getters.enabledAuth && $store.getters.role === 1" class="item" to="/admin/account/list" v-bind:class="{active: this.$route.path.indexOf('/admin/account') === 0}"><i class="user icon"></i> {{$L('account')}}</router-link>

      <div class="right menu">
        <router-link to="/user/setpwd" class="item" v-if="this.$store.getters.email"><i class="user icon"></i> {{this.$store.getters.email}}</router-link>
        <a class="item" v-if="this.$store.getters.email" href="#" v-on:click="logout"><i class="sign out icon"></i></a>

        <div ref="langSelection" class="ui right icon dropdown item">
          <i class="world icon" style="margin-left:-1px; margin-right: 8px;"></i>
          <span class="text">Language</span>
          <i class="dropdown icon"></i>
          <div class="menu">
            <div class="item" v-for="lang in $Lang.supported" :data-value="lang.code">{{lang.name}}</div>
          </div>
        </div>
      </div>
    </div>
    <div style="height: 55px;"></div>
    <div class="ui container">
      <router-view></router-view>
    </div>
    <Messager/>
  </div>
</template>

<script>
import store from './vuex/store';
import Vue from 'vue';
import Messager from './components/Messager.vue';

export default {
  name: 'app',
  store,

  mounted: function(){
    $(this.$refs.langSelection).dropdown({
      onChange: function(value, text){
        var old = window.$.cookie('locale');
        if (old !== value) {
          window.$.cookie('locale', value)
          window.location.reload()
        }
      }
    });
  },

  computed: {
    shouldOpen() {
      return !this.$store.getters.enabledAuth || (this.$store.getters.enabledAuth && this.$store.getters.email)
    }
  },

  methods: {
    logout() {
      var vm = this;
      this.$rest.DELETE('session').
        onsucceed(200, ()=>{
          vm.$store.commit('setEmail', '');
          vm.$store.commit('setRole', 0);
          vm.$router.push('/login');
        }).
        do();
    }
  },

  components: {
    Messager
  },
}
</script>
