<template>
<div style="margin-bottom:20px;">
  <h4 class="ui horizontal divider header">定时器 - {{index}} <a href="#" v-on:click.prevent="remove">删除</a></h4>
  <div class="two fields">
    <div class="field">
      <input type="text" v-bind:value="rule.timer" v-on:input="change('timer', $event.target.value)" placeholder="定时 * * * * *（crontab 格式）"/>
    </div>
    <div class="field">
      <Dropdown title="节点分组" v-bind:items="nodeGroups" multiple="true"></Dropdown>
    </div>
  </div>
  <div class="field">
    <label><strong style="color:green;">+</strong> 同时在这些节点上面运行任务</label>
    <Dropdown title="选择节点" v-bind:items="activityNodes" multiple="true"></Dropdown>
  </div>
  <div class="field">
    <label><strong style="color:red;">-</strong> 不在这些节点上面运行任务</label>
    <Dropdown title="选择节点" v-bind:items="activityNodes" multiple="true"></Dropdown>
  </div>
</div>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';

export default {
  name: 'job-edit-rule',
  props: ['rule', 'index'],
  data: function(){
    return {
      nodeGroups: [],
      activityNodes: []
    }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('node/activitys').onsucceed(200, (resp)=>{vm.activityNodes = resp}).do();
    this.$rest.GET('node/groups').onsucceed(200, (resp)=>{
      var groups = [];
      for (var i in resp) {
        groups.push({value: resp[i].id, name: resp[i].name});
      }
      vm.nodeGroups = groups;
    });
  },

  methods: {
    remove: function(){
      this.$emit('remove', this.index);
    },
    change: function(key, val){
      this.$emit('change', this.index, key, val);
    }
  },

  components: {
    Dropdown
  }
}
</script>