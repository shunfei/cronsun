<template>
  <div v-if="error != ''" class="ui negative message">
    <div class="header"><i class="attention icon"></i> {{error}}</div>
  </div>
  <form v-else class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <h3 class="ui header">{{action == 'CREATE' ? '添加' : '更新'}}节点分组</h3>
    <div class="field">
      <input type="text" ref="name" v-model:value="group.name" placeholder="分组名称">
    </div>
    <div class="field">
      <label>分组节点</label>
      <Dropdown title="选择节点" multiple="true" v-bind:items="allNodes" v-bind:selected="group.nids" v-on:change="changeGroup"/>
    </div>
    <div class="field">
      <button class="fluid blue ui button" type="button" v-on:click="submit"><i class="upload icon"></i> 保存分组</button>
    </div>
    <div class="field">
      <button class="fluid red ui button" type="button" v-on:click="remove"><i class="remove icon"></i> 删除分组</button>
    </div>
  </form>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';

export default {
  name: 'node_group_edit',
  data(){
    return {
      error: '',
      loading: false,
      action: '',
      allNodes: [],
      group: {
        id: '',
        name: '',
        nids: ''
      }
    }
  },

  mounted(){
    var vm = this;

    if (this.$route.path.indexOf('/node/group/create') === 0) {
      this.action = 'CREATE';
    } else {
      this.action = 'UPDATE';
      this.loading = true;
      this.$rest.GET('node/group/'+this.$route.params.id).
        onsucceed(200, (resp)=>{
          vm.group = resp;
        }).
        onfailed((data)=>{vm.error = data.error}).
        onend(()=>{vm.loading = false}).
        do();
    }

    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      var allNodes = [];
      for (var i in resp) {
        allNodes.push(resp[i].id);
      }
      vm.allNodes = allNodes;
    }).do();
  },

  methods: {
    changeGroup(val, text){
      if (val.length == 0) {
        this.group.nids = [];
        return;
      }
      this.group.nids = val.split(',');
    },

    submit(){
      var exceptCode = this.action == 'CREATE' ? 201 : 200;
      this.loading = true;
      var vm = this;
      this.$rest.PUT('node/group', this.group)
        .onsucceed(exceptCode, ()=>{vm.$router.push('/node/group')})
        .onfailed((resp)=>{console.log(resp)})
        .onend(()=>{vm.loading=false})
        .do();
    },

    remove(){
      if (!confirm('确定删除该分组 ' + this.group.name + '?')) return;
      var vm = this;
      this.$rest.DELETE('node/group/'+this.group.id)
        .onsucceed(204, ()=>{vm.$router.push('/node/group')})
        .onfailed((resp)=>{console.log(resp)})
        .onend(()=>{vm.loading=false})
        .do();
    }
  },
  
  components: {
    Dropdown
  }
}
</script>