<template>
  <div id="app">
    <div class="ui purple inverted menu fixed">
      <router-link class="active item" to="/">仪表盘</router-link>
      <router-link class="item" to="/log">日志</router-link>
      <router-link class="item" to="/job">任务</router-link>
      <router-link class="item" to="/node">节点</router-link>
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