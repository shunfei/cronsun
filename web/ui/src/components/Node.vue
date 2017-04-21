<style scoped>
.node {
  width: 130px;
  border-radius: 3px;
  padding: 4px 0;
  margin: 3px;
  display: inline-block;
  background: #e8e8e8;
  text-align: center;
}
.notice {
  color: red;
}
</style>
<template>
  <div>
    <div class="clearfix">
      <router-link class="ui right floated primary button" to="/node/group"><i class="cubes icon"></i> 管理分组</router-link>
      <div class="ui label" title="正常运行的节点"><i class="green cube icon"></i> {{items[2].nodes.length}} 正常节点</div>
      <div class="ui label" title="手动下线/维护中的"><i class="cube icon"></i> {{items[1].nodes.length}} 离线节点</div>
      <div class="ui label" title="因自身或网络等原因未检测到节点存活"><i class="red cube icon"></i> {{items[0].nodes.length}} 故障节点</div>
      （总 {{count}} 个节点）
      <div class="ui label" title="当前版本号"> {{version}} </div>
    </div>
    <div class="ui relaxed list" v-for="item in items">
      <h4 v-if="item.nodes.length > 0" class="ui horizontal divider header"><i class="cube icon" v-bind:class="item.css"></i> {{item.name}} {{item.nodes.length}}</h4>
      <div v-for="node in item.nodes" class="node" v-bind:class="[(node.version == version) ? '' : 'notice']" v-bind:title="node.title">{{node.id}}</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'node',
  data: function(){
    return {
      items: [
        {nodes:[],name:'故障节点',css:'red'},
        {nodes:[],name:'离线节点',css:''},
        {nodes:[],name:'正常节点',css:'green'}
      ],
      count: 0,
      version: ''
    }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('version').onsucceed(200, (resp)=>{
      vm.version = resp;
    }).do();
    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      resp.sort(function(a, b){
        var aid = a.id.split('.');
        var bid = b.id.split('.');
        var ai = 0, bi = 0;
        for (var i in aid) {
          ai += (+aid[i])*Math.pow(255,3-i);
          bi += (+bid[i])*Math.pow(255,3-i);
        }
        return ai - bi;
      });
      for (var i in resp) {
        var n = resp[i];
        n.title = n.version + "\nstarted at: " + n.up
        if (n.alived && n.connected) {
          vm.items[2].nodes.push(n);
        } else if (n.alived && !n.connected) {
          vm.items[0].nodes.push(n);
        } else {
          vm.items[1].nodes.push(n);
        }
      }
      vm.count = resp.length || 0;
    }).do();
  }
}
</script>