<template>
  <div>
    <h3 class="ui horizontal divider header">当前节点： {{count}}</h3>
    <div v-for="node in nodes" v-bind:class="{blue:node.alived}" class="ui label"><i class="cube icon"></i> {{node.id}}</div>
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
    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      vm.nodes = resp;
      vm.count = resp.length;
    }).do();
  }
}
</script>