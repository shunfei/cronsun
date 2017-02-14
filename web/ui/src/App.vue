<template>
  <div id="app">
    <div class="ui blue inverted menu fixed">
      <div class="item">CRONSUN</div>
      <router-link class="item" to="/" v-bind:class="{active: this.$route.path == '/'}"><i class="dashboard icon"></i> 仪表盘</router-link>
      <router-link class="item" to="/log" v-bind:class="{active: this.$route.path.indexOf('/log') === 0}"><i class="file text icon"></i> 日志</router-link>
      <router-link class="item" to="/job" v-bind:class="{active: this.$route.path.indexOf('/job') === 0}"><i class="calendar icon"></i> 任务</router-link>
      <router-link class="item" to="/nodes" v-bind:class="{active: this.$route.path.indexOf('/nodes') === 0}"><i class="server icon"></i> 节点</router-link>
    </div>
    <div style="height: 55px;"></div>
    <div class="ui container">
      <router-view></router-view>
    </div>
  </div>
</template>

<script>
export default {
  name: 'app'
}
</script>