<template>
  <div>
    <h3 class="ui horizontal divider header">当前活动节点： {{count}}</h3>
    <div class="ui label" v-for="node in nodes"><i class="cube icon"></i> {{node}}</div>
  </div>
</template>

<script>
export default {
  name: 'node',
  data: function(){
    return {
      count: 0,
      nodes: []
    }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('node/activitys').onsucceed(200, (resp)=>{
      vm.nodes = resp;
      vm.count = resp.length;
    }).do();
  }
}
</script>