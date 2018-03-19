<style scoped>
.ui.vertical.menu h4 {
  background: rgba(0,0,0,.05);
  margin-bottom: 0px;
  margin-top: 0px;
}
</style>

<template>
  <div id="app">
    <div class="ui blue inverted menu fixed">
      <div class="item">CRONSUN</div>
      <router-link v-if="shouldOpen" class="item" to="/" v-bind:class="{active: this.$route.path == '/'}"><i class="dashboard icon"></i> {{$L('dashboard')}}</router-link>
      <router-link v-if="shouldOpen" class="item" to="/log" v-bind:class="{active: this.$route.path.indexOf('/log') === 0}"><i class="file text icon"></i> {{$L('log')}}</router-link>
      <router-link v-if="shouldOpen" class="item" to="/job" v-bind:class="{active: this.$route.path.indexOf('/job') === 0}"><i class="calendar icon"></i> {{$L('job')}}</router-link>
      <router-link v-if="shouldOpen" class="item" to="/node" v-bind:class="{active: this.$route.path.indexOf('/node') === 0}"><i class="server icon"></i> {{$L('node')}}</router-link>
      <router-link v-if="$store.getters.enabledAuth && $store.getters.role === 1" class="item" to="/admin/account/list" v-bind:class="{active: this.$route.path.indexOf('/admin/account') === 0}"><i class="user icon"></i> {{$L('account')}}</router-link>
      <a class="item" href="https://github.com/shunfei/cronsun/wiki" target="_blank"><i class="external alternate icon"></i> Docs</a>
      <div class="right menu">
        <router-link to="/user/setpwd" class="item" v-if="this.$store.getters.email"><i class="user icon"></i> {{this.$store.getters.email}}</router-link>
        <a class="item" v-if="this.$store.getters.email" href="#" v-on:click="logout"><i class="sign out icon"></i></a>

        <a class="item" href="#" @click.prevent="toggleSetting"><i class="settings icon"></i></a>
      </div>
    </div>
    <div style="height: 55px;"></div>
    <div class="ui container pusher">
      <router-view></router-view>
    </div>

    <div ref="sidebar" class="ui sidebar right vertical menu">
      <h4 class="item">{{$store.getters.version}}</h4>
      <h4 class="item">Language</h4>
      <div class="item">
        <div class="menu">
          <a class="item" :class="{active: locale === lang.code}" href="#"  v-for="lang in $Lang.supported" @click.prevent="selectLanguage(lang.code)">{{lang.name}}</a>
        </div>
      </div>
      <h4 class="item">{{$L('node show as')}}</h4>
      <div class="item">
        <div class="menu">
          <a class="item" :class="{active: hostshow === h.value}" href="#" v-for="h in hostshowsList" @click.prevent="selectHostshows(h.value)">{{h.name}}</a>
        </div>
      </div>
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

  data: function() {
    return {
      hostshowsList: [{name: this.$L('hostname'), value: 'hostname'}, {name: 'IP', value: 'ip'}],
      locale: ''
    }
  },

  created: function() {
    this.locale = window.$.cookie('locale');
    this.$store.commit('setShowWithHostname', window.$.cookie('hostshows') === 'hostname');
  },

  computed: {
    shouldOpen() {
      return !this.$store.getters.enabledAuth || (this.$store.getters.enabledAuth && this.$store.getters.email)
    },

    hostshow() {
      return this.$store.getters.showWithHostname ? 'hostname' : 'ip';
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
    },

    toggleSetting() {
      $(this.$refs.sidebar).sidebar('toggle');
    },

    selectLanguage(code) {
      if (this.locale !== code) {
        window.$.cookie('locale', code);
        window.location.reload();
        return;
      }
      this.toggleSetting();
    },

    selectHostshows(v) {
      if (this.hostshow !== v) {
        window.$.cookie('hostshows', v);
        window.location.reload();
        return;
      }
      this.toggleSetting();
    }
  },

  components: {
    Messager
  },
}
</script>
