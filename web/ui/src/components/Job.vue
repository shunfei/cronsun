<style scope>
  .clearfix:after {content:""; clear:both; display:table;}
  .ui.fitted.checkbox {min-height: 15px;}
</style>
<template>
  <div>
    <div class="clearfix" style="margin-bottom: 20px;">
      <router-link class="ui left floated button" to="/job/executing">{{$L('view executing jobs')}}</router-link>
      <button class="ui left floated icon button" v-on:click="refresh"><i class="refresh icon"></i></button>
      <div class="ui icon buttons">
        <button class="ui left floated icon button" v-on:click="batched=!batched">{{$L('batch')}}</button>
        <button class="ui button" :class="{disabled: batchIds.length == 0}" v-if="batched" v-on:click="batch('start')">
          <i class="play icon"></i>
        </button>
        <button class="ui button" :class="{disabled: batchIds.length == 0}" v-if="batched" v-on:click="batch('pause')">
          <i class="pause icon"></i>
        </button>
      </div>
      <router-link class="ui right floated primary button" to="/job/create"><i class="add to calendar icon"></i> {{$L('create job')}}</router-link>
    </div>
    <form class="ui form">
      <div class="two fields">
        <div class="field">
          <label>{{$L('group filter')}}</label>
          <Dropdown :title="$L('select a group')" v-bind:items="groups" v-on:change="changeGroup" :selected="group"/>
        </div>
        <div class="field">
          <label>{{$L('node filter')}}</label>
          <Dropdown :title="$L('select a node')" v-bind:items="nodes" v-on:change="changeNode" :selected="node"/>
        </div>
      </div>
    </form>
    <table class="ui hover blue table" v-if="jobs.length > 0">
      <thead>
        <tr>
          <th class="collapsing center aligned">{{$L('operation')}}</th>
          <th class="collapsing center aligned">{{$L('status')}}</th>
          <th width="200px" class="center aligned">{{$L('group')}}</th>
          <th class="center aligned">{{$L('user')}}</th>
          <th class="center aligned">{{$L('name')}}</th>
          <th class="center aligned">{{$L('latest executed')}}</th>
          <th class="center aligned">{{$L('executing result')}}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(job, index) in jobs">
          <td class="center aligned">
            <div class="ui icon dropdown" v-show="!batched">
              <i class="content icon"></i>
              <div class="menu">
                <div class="item" v-on:click="$router.push('/job/edit/'+job.group+'/'+job.id)">{{$L('edit')}}</div>
                <div class="item" v-if="job.pause" v-on:click="changeStatus(job.group, job.id, index, !job.pause)">{{$L('open')}}</div>
                <div class="item" v-if="!job.pause" v-on:click="changeStatus(job.group, job.id, index, !job.pause)">{{$L('pause')}}</div>
                <div class="divider"></div>
                <div class="item" style="color:red;" v-on:click="removeJob(job.group, job.id, index)">{{$L('delete')}}</div>
              </div>
            </div>
            <div class="ui fitted checkbox" v-show="batched">
              <input type="checkbox" :value="job.group+'/'+job.id" v-model="batchIds"><label></label>
            </div>
          </td>
          <td class="center aligned"><i class="icon" v-bind:class="{pause: job.pause, play: !job.pause, green: !job.pause}"></i></td>
          <td>{{job.group}}</td>
          <td>{{job.user}}</td>
          <td><router-link :to="'/job/edit/'+job.group+'/'+job.id">{{job.name}}</router-link></td>
          <td>
            <span v-if="!job.latestStatus">-</span>
            <span v-else>{{formatLatest(job.latestStatus)}}</span>
          </td>
          <td :class="{error: job.latestStatus && !job.latestStatus.success}">
            <span v-if="!job.latestStatus">-</span>
            <router-link v-else :to="'/log/'+job.latestStatus.refLogId">{{$L(job.latestStatus.success ? 'successed' : 'failed')}}</router-link> |
            <router-link :to="{path: 'log', query: {latest:true, ids: job.id}}">latest</router-link> |
            <a href="#" :title="$L('click to select a node and re-execute job')" v-on:click.prevent="showExecuteJobModal(job.name, job.group, job.id)"><i class="icon repeat"></i></a>
          </td>
        </tr>
      </tbody>
    </table>
    <ExecuteJob ref="executeJobModal"/>
  </div>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';
import ExecuteJob from './ExecuteJob.vue';
import {formatTime, formatDuration} from '../libraries/functions';

export default {
  name: 'job',
  data: function(){
    return {
      batched: false,
      batchIds: [],
      groups: [],
      group: '',
      nodes: [],
      node: '',
      jobs: []
    }
  },
  
  mounted: function(){
    this.fillParams();
    var vm = this;

    this.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      resp.unshift({value: '', name: vm.$L('all groups')});
      vm.groups = resp;
      this.fetchList(this.buildQuery());
    }).do();

    this.$rest.GET('nodes').onsucceed(200, (resp)=>{
      vm.nodes.push({name: vm.$L('all nodes'), value: ''});
      for (var i in resp) {
        vm.nodes.push(resp[i].id);
      }
    }).do();

    $('.ui.checkbox').checkbox();
  },

  watch: {
    '$route': function(){
      this.fillParams();
      this.fetchList(this.buildQuery());
    }
  },

  methods: {
    fillParams: function(){
      this.group = this.$route.query.group || '';
      this.node = this.$route.query.node || '';
    },

    changeGroup: function(val, text){
      var vm = this;
      this.group = val;
      this.$router.push('job?'+this.buildQuery());
    },

    changeNode: function(val, text){
      var vm = this;
      this.node = val;
      this.$router.push('job?'+this.buildQuery());
    },

    buildQuery: function(){
      var params = [];
      if (this.group) params.push('group='+this.group);
      if (this.node) params.push('node='+this.node);
      return params.join('&');
    },

    fetchList: function(query){
      var vm = this;
      this.$rest.GET('jobs?'+query).onsucceed(200, (resp)=>{
        vm.jobs = resp;
        vm.$nextTick(()=>{
          $(vm.$el).find('table .ui.dropdown').dropdown();
        });
      }).do();
    },

    refresh: function(){
      this.fetchList(this.buildQuery());
    },

    removeJob: function(group, id, index){
      var vm = this;
      this.$rest.DELETE('job/'+group+'-'+id).onsucceed(204, (resp)=>{
        vm.jobs.splice(index, 1);
      }).do();
    },

    changeStatus: function(group, id, index, isPause){
      var vm = this;
      this.$rest.POST('job/'+group+'-'+id, {"pause": isPause}).onsucceed(200, (resp)=>{
        vm.refresh();
      }).do();
    },

    formatExecResult: function(st){
      if (!st) return '-';
      return 
    },

    formatLatest: function(latest){
      return this.$L('on {node} took {times}, {begin ~ end}', latest.node, formatDuration(latest.beginTime, latest.endTime), formatTime(latest.beginTime, latest.endTime));
    },

    showExecuteJobModal: function(jobName, jobGroup, jobId){
      this.$refs.executeJobModal.show(jobName, jobGroup, jobId);
    },

    batch: function(op){
      switch(op) {
        case 'start': break;
        case 'pause': break;
        default: return;
      }

      var vm = this;
      this.$rest.POST('jobs/'+op, this.batchIds).onsucceed(200, (resp)=>{
        vm.refresh();
        vm.$bus.$emit('warning', resp);
      }).do();
    }
  },

  components: {
    Dropdown,
    ExecuteJob
  }
}
</script>
