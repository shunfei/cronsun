<style scope>
 #nodeGroups {
    margin: 15px 0;
    position: relative;
    column-width: 250px;
}
#nodeGroups>div {
    width: 100%;
    display: inline-block;
    border: 1px solid #D4D4D5; box-shadow: none; margin: 0px;
    margin-bottom: 1em;
}

#nodeGroups .ui.card>.content {padding: 1em 0 0;}
#nodeGroups .ui.list:first-child {border-top: 1px solid #eee;}
#nodeGroups .ui.divided.list>.item {padding: 0.5em 1em;}
#nodeGroups .ui.card>.content>.header:not(.ui) {margin: 0 1em;}
</style>

<template>
  <div>
    <div class="clearfix">
      <router-link class="ui right floated primary button" to="/node/group/create"><i class="add icon"></i> {{$L('create group')}}</router-link>
      <button class="ui right floated icon button" v-on:click="refresh"><i class="refresh icon"></i></button>
    </div>
    <div v-if="error != ''" class="header"><i class="attention icon"></i> {{error}}</div>
    <div id="nodeGroups">
      <div class="ui card" v-for="g in groups">
        <div class="content">
          <router-link class="header" :to="'/node/group/'+g.id">{{g.name}}</router-link>
          <div class="description">
            <div class="ui middle large aligned divided list"> 
              <div class="item" v-for="nodeID in g.nids">
                <span v-if="nodes[nodeID]">{{$store.getters.hostshows(nodeID)}}
                <i class="arrow circle up icon red" v-if="nodes[nodeID].hostname == ''"></i>
                <i v-if="nodes[nodeID].hostname == ''">(need to upgrade)</i>
                </span>
                <span v-else :title="$L('node not found, was it removed?')">{{nodeID}} <i class="question circle icon red"></i></span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'node_group',
  data: function(){
    return {
      error: '',
      groups: []
    }
  },
  mounted: function(){
    this.refresh();
  },

  methods: {
    refresh(){
      var vm = this;
      this.$rest.GET('node/groups').
        onsucceed(200, (resp)=>{
          vm.groups = resp;
        }).
        onfailed((data)=>{vm.error = data}).
        onend(()=>{vm.loading = false}).
        do();
    }
  },

  computed: {
    nodes: function () {
      return this.$store.getters.nodes;
    }
  }
}
</script>
