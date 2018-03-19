<style scoped>
.node {
  padding: 0 13px;
  border-radius: 3px;
  margin: 3px;
  display: inline-block;
  background: #e8e8e8;
  text-align: center;
  position: relative;

  line-height: 1.9em;
}

.node > i.icon.remove {
  position: absolute;
  top: 0;
  right: 0;
  background: #2185D0;
  bottom: 0;
  display: none;
  color: white;
  margin: 0;
  height: auto;
  width: 100%;
  cursor: pointer;
}
.node:hover > i.icon.remove {display: block;}
</style>
<template>
  <div>
    <div class="clearfix">
      <router-link class="ui right floated primary button" to="/node/group"><i class="cubes icon"></i> {{$L('group manager')}}</router-link>
      <div class="ui label" v-for="group in groups" v-bind:title="$L(group.title)">
        <i class="cube icon" v-bind:class="group.css"></i> {{group.nodes.length}} {{$L(group.name)}}
      </div>
      {{$L('(total {n} nodes)', count)}}
      <div class="ui label" :title="$L('currently version')"> {{version}}</div>
    </div>
    <div class="ui relaxed list" v-for="(group, groupIndex) in groups">
      <h4 v-if="group.nodes.length > 0" class="ui horizontal divider header"><i class="cube icon" v-bind:class="group.css"></i> {{$L(group.name)}} {{group.nodes.length}}</h4>
      <div v-for="(node, nodeIndex) in group.nodes" class="node" v-bind:title="node.title">
        <router-link class="item" :to="'/job?node='+node.id">
          <i class="red icon fork" v-if="node.version !== version" :title="$L('version inconsistent, node: {version}', node.version)"></i>
          {{$store.getters.hostshows(node.id)}}
        </router-link>
        <i v-if="groupIndex != 2" v-on:click="removeConfirm(groupIndex, nodeIndex, node.id)" class="icon remove"></i>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'node',
  data: function(){
    return {
      groups: [
        {nodes: [], name: 'node damaged', title: 'node can not be deceted due to itself or network etc.', css:'red'},
        {nodes: [], name: 'node offline', title: 'node is in maintenance or is shutdown manually', css:''},
        {nodes: [], name: 'node normaly', title: 'node is running', css:'green'}
      ],
      count: 0
    }
  },

  mounted: function(){
    var vm = this;

    var nodes = this.$store.getters.nodes;
    for (var id in nodes) {
      var n = nodes[id];
      n.title = n.ip + "\n" + n.id + "\n" + n.version + "\nstarted at: " + n.up
      if (n.alived && n.connected) {
        vm.groups[2].nodes.push(n);
      } else if (n.alived && !n.connected) {
        vm.groups[0].nodes.push(n);
      } else {
        vm.groups[1].nodes.push(n);
      }
    }

    vm.count = Object.keys(nodes).length;
  },

  computed: {
    version: function(){
      return this.$store.getters.version;
    }
  },

  methods: {
    removeConfirm: function(groupIndex, nodeIndex, nodeId){
      if (!confirm(this.$L('are you sure to remove the node {nodeId}?', nodeId))) return;
      
      var vm = this;
      this.$rest.DELETE('node/'+nodeId).onsucceed(204, (resp) => {
        vm.groups[groupIndex].nodes.splice(nodeIndex, 1);
      }).do();
    }
  }
}
</script>
