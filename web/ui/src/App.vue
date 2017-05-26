<template>
  <div id="app">
    <div class="ui blue inverted menu fixed">
      <div class="item">CRONSUN</div>
      <router-link class="item" to="/" v-bind:class="{active: this.$route.path == '/'}"><i class="dashboard icon"></i> {{$L('dashboard')}}</router-link>
      <router-link class="item" to="/log" v-bind:class="{active: this.$route.path.indexOf('/log') === 0}"><i class="file text icon"></i> {{$L('log')}}</router-link>
      <router-link class="item" to="/job" v-bind:class="{active: this.$route.path.indexOf('/job') === 0}"><i class="calendar icon"></i> {{$L('job')}}</router-link>
      <router-link class="item" to="/node" v-bind:class="{active: this.$route.path.indexOf('/node') === 0}"><i class="server icon"></i> {{$L('node')}}</router-link>

      <div ref="langSelection" class="ui right icon dropdown item">
        <i class="world icon" style="margin-left:-1px; margin-right: 8px;"></i>
        <span class="text">Language</span>
        <i class="dropdown icon"></i>
        <div class="menu">
          <div class="item" v-for="lang in $Lang.supported" :data-value="lang.code">{{lang.name}}</div>
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
import Messager from './components/Messager.vue';

export default {
  name: 'app',

  mounted: function(){
    var vm = this;
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

  components: {
    Messager
  },
}
</script>