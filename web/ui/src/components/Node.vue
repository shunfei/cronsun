<template>
  <div>
    <h3 class="ui horizontal divider header">当前节点： {{count}}</h3>
    <div class="ui label" title="手动下线/维护中的"><i class="cube icon"></i> 已下线节点</div>
    <div class="ui label" title="正常运行的节点"><i class="green cube icon"></i> 正常节点</div>
    <div class="ui label" title="节点因自身或网络等原因未检测到"><i class="red cube icon"></i> 故障节点</div>
    <h3 class="ui horizontal divider header"> </h3>
    <div v-for="node in nodes" class="ui label"><i v-bind:class="{green: node.alived && node.connected, red: node.alived && !node.connected}" class="cube icon"></i> {{node.id}}</div
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