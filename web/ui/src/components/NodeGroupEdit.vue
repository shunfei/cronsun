<template>
  <div v-if="error != ''" class="ui negative message">
    <div class="header"><i class="attention icon"></i> {{error}}</div>
  </div>
  <form v-else class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
    <h3 class="ui header">{{$L((action == 'CREATE' ? 'create' : 'update')+' node group')}}</h3>
    <div class="field">
      <input type="text" ref="name" v-model:value="group.name" :placeholder="$L('group name')">
    </div>
    <div class="field">
      <label>{{$L('include nodes')}}</label>
      <Dropdown :title="$L('select nodes')" multiple="true" v-bind:items="$store.getters.dropdownNodes" v-bind:selected="group.nids" v-on:change="changeGroup"/>
    </div>
    <div class="field">
      <button class="fluid blue ui button" type="button" v-on:click="submit"><i class="upload icon"></i> {{$L('save group')}}</button>
    </div>
    <div class="field">
      <button class="fluid red ui button" type="button" v-on:click="remove"><i class="remove icon"></i> {{$L('delete group')}}</button>
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
      if (!confirm(this.$L('are you sure to delete the group {name}?', this.group.name))) return;
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
