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
</style>
<template>
  <div>
    <div class="clearfix">
      <router-link class="ui right floated primary button" to="/node/group"><i class="cubes icon"></i> {{$L('group manager')}}</router-link>
      <div class="ui label" 
      <div class="ui label" v-for="item in items" v-bind:title="$L(item.title)">
        <i class="cube icon" v-bind:class="item.css"></i> {{item.nodes.length}} {{$L(item.name)}}
      </div>
      {{$L('(total {n} nodes)', count)}}
      <div class="ui label" :title="$L('currently version')"> {{version}} </div>
    </div>
    <div class="ui relaxed list" v-for="item in items">
      <h4 v-if="item.nodes.length > 0" class="ui horizontal divider header"><i class="cube icon" v-bind:class="item.css"></i> {{$L(item.name)}} {{item.nodes.length}}</h4>
      <div v-for="node in item.nodes" class="node" v-bind:title="node.title">
        <router-link class="item" :to="'/job?node='+node.id">
          <i class="red icon fork" v-if="node.version !== version" :title="$L('version inconsistent, node: {version}', node.version)"></i>
          {{node.id}}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'node',
  data: function(){
    return {
      items: [
        {nodes:[],name:'node damaged',title:'node can not be deceted due to itself or network etc.',css:'red'},
        {nodes:[],name:'node offline',title:'node is in maintenance or is shutdown manually',css:''},
        {nodes:[],name:'node normaly',title:'node is running',css:'green'}
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