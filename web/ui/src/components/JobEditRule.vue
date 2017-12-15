<template>
<div style="margin-bottom:20px;">
  <h4 class="ui horizontal divider header">{{$L('timer')}} - {{index}} <a href="#" v-on:click.prevent="remove">{{$L('delete')}}</a></h4>
  <div class="two fields">
    <div class="field">
      <div class="ui icon input">
        <input type="text" v-bind:value="rule.timer" v-on:input="change('timer', $event.target.value)" :placeholder="$L('0 * * * * *, rules see the 「?」on the right')"/>
        <i ref="ruletip" class="large help circle link icon" data-position="top right" :data-content="$L('<sec> <min> <hr> <day> <month> <week>, rules is same with Cron')" data-variation="wide"></i>
      </div>
    </div>
    <div class="field">
      <Dropdown :title="$L('node group')" v-bind:items="nodeGroups" v-bind:selected="rule.gids" multiple="true" v-on:change="changeNodeGroups($event)"></Dropdown>
    </div>
  </div>
  <div class="field">
    <label>{{$L('and please running on those nodes')}}</label>
    <Dropdown :title="$L('select nodes')" v-bind:items="activityNodes" v-bind:selected="rule.nids" v-on:change="changeIncludeNodes($event)" multiple="true"></Dropdown>
  </div>
  <div class="field">
    <label>{{$L('do not running on those nodes')}}</label>
    <Dropdown :title="$L('select nodes')" v-bind:items="activityNodes" v-bind:selected="rule.exclude_nids" v-on:change="changeExcludeNodes($event)" multiple="true"></Dropdown>
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


    this.$rest.GET('node/groups').onsucceed(200, (resp)=>{
      var groups = [];
      for (var i in resp) {
        groups.push({value: resp[i].id, name: resp[i].name});
      }
      vm.nodeGroups = groups;
    }).do();

    $(this.$refs.ruletip).popup();
  },

  methods: {
    remove: function(){
      this.$emit('remove', this.index);
    },
    change: function(key, val){
      this.$emit('change', this.index, key, val);
    },
    changeNodeGroups: function(val){
      var gids = val.trim().length === 0 ? [] : val.split(',');
      this.change('gids', gids);
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
