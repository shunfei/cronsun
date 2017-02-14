<template>
<div style="margin-bottom:20px;">
  <h4 class="ui horizontal divider header">定时器 - {{index}} <a href="#" v-on:click.prevent="remove">删除</a></h4>
  <div class="two fields">
    <div class="field">
      <div class="ui icon input">
        <input type="text" v-bind:value="rule.timer" v-on:input="change('timer', $event.target.value)" placeholder="定时 * 5 * * * *"/>
        <i ref="ruletip" class="large help circle link icon" data-position="top right" data-content="<秒> <分钟> <小时> <日> <月份> <星期>，规则与 crontab 一样" data-variation="wide"></i>
      </div>
    </div>
    <div class="field">
      <Dropdown title="节点分组" v-bind:items="nodeGroups" multiple="true"></Dropdown>
    </div>
  </div>
  <div class="field">
    <label>同时在这些节点上面运行任务</label>
    <Dropdown title="选择节点" v-bind:items="activityNodes" v-bind:selected="rule.nids" v-on:change="changeIncludeNodes($event)" multiple="true"></Dropdown>
  </div>
  <div class="field">
    <label>不在这些节点上面运行任务</label>
    <Dropdown title="选择节点" v-bind:items="activityNodes" v-on:change="changeExcludeNodes($event)" multiple="true"></Dropdown>
  </div>
</div>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';

export default {
  name: 'job-edit-rule',
  props: ['index', 'rule'],
  data: function(){
    return {
      nodeGroups: [],
      activityNodes: []
    }
  },

  mounted: function(){
    var vm = this;
    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      for (var i in resp) {
        vm.activityNodes.push(resp[i].id);
      }
    }).do();
    this.$rest.GET('nodes/groups').onsucceed(200, (resp)=>{
      var groups = [];
      for (var i in resp) {
        groups.push({value: resp[i].id, name: resp[i].name});
      }
      vm.nodeGroups = groups;
    });

    $(this.$refs.ruletip).popup();
  },

  methods: {
    remove: function(){
      this.$emit('remove', this.index);
    },
    change: function(key, val){
      this.$emit('change', this.index, key, val);
    },
    changeIncludeNodes: function(val){
      var nids = val.trim().length === 0 ? [] : val.split(',');
      this.change('nids', nids);
    },
    changeExcludeNodes: function(val){
      var exclude_nids = val.trim().length === 0 ? [] : val.split(',');
      this.change('exclude_nids', exclude_nids);
    }
  },

  components: {
    Dropdown
  }
}
</script>