<template>
  <div>
    <form class="ui form" method="GET" v-bind:class="{loading:loading}" v-on:submit.prevent>
      <div class="two fields">
        <div class="field">
          <label>任务名称</label>
          <input type="text" v-model="names"  placeholder="多个名称用英文逗号分隔">
        </div>
        <div class="field">
          <label>任务 ID</label>
          <input type="text" v-model="ids"  placeholder="多个 ID 用英文逗号分隔">
        </div>
      </div>
      <div class="field">
        <label>运行节点</label>
        <input type="text" v-model="nodes" placeholder="ip，多个 ip 用英文逗号分隔">
      </div>
      <div class="two fields">
        <div class="field">
          <label>开始时间</label>
          <input type="date" v-model="begin">
        </div>
        <div class="field">
          <label>结束时间</label>
          <input type="date" v-model="end">
        </div>
      </div>
      <div class="two fields">
        <div class="filed">
          <div ref="latest" class="ui checkbox">
            <input type="checkbox" class="hidden" v-model="latest">
            <label>只看每个任务在每个节点上最后一次运行的结果</label>
          </div>
        <div class="filed">
          <div ref="failedOnly" class="ui checkbox">
            <input type="checkbox" class="hidden" v-model="failedOnly">
            <label>只看失败的任务</label>
          </div>
        </div>
        </div>
      </div>
      <div class="field">
        <button class="fluid ui button" type="button" v-on:click="submit">查询</button>
      </div>
    </form>
    <table class="ui selectable green table" v-if="list && list.length > 0">
      <thead>
        <tr>
          <th class="center aligned">任务名称</th>
          <th class="center aligned">运行节点</th>
          <th class="center aligned">执行用户</th>
          <th class="center aligned">执行时间</th>
          <th class="center aligned">运行结果</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="log in list">
          <td><router-link class="item" :to="'/job/edit/'+log.jobGroup+'/'+log.jobId">{{log.name}}</router-link></td>
          <td>{{log.node}}</td>
          <td>{{log.user}}</td>
          <td :class="{warning: durationAttention(log.beginTime, log.endTime)}"><i class="attention icon" v-if="durationAttention(log.beginTime, log.endTime)"></i> {{formatTime(log)}}</td>
          <td :class="{error: !log.success}">
            <router-link :to="'/log/'+log.id">{{log.success ? '成功' : '失败'}}</router-link> |
            <a href="#" title="点此选择节点重新执行任务" v-on:click.prevent="showExecuteJobModal(log.name, log.jobGroup, log.jobId)"><i class="icon repeat"></i></a>
          </td>
        </tr>
      </tbody>
    </table>
    <Pager v-if="list && list.length>0" :total="total" :length="5"/>
    <ExecuteJob ref="executeJobModal"/>
  </div>
</template>

<script>
import Pager from './basic/Pager.vue';
import ExecuteJob from './ExecuteJob.vue';
import {formatTime, formatDuration} from '../libraries/functions';

export default {
  name: 'log',
  data: function(){
    return {
      loading: false,
      names: '',
      ids: '',
      nodes: '',
      begin: '',
      end: '',
      latest: false,
      failedOnly: '',
      list: [],
      total: 0,
      page: 1
    }
  },

  mounted: function(to, from, next){
      this.fillParams();
      this.fetchList(this.buildQuery());

      var vm = this;
      $(this.$refs.latest).checkbox({'onChange': ()=>{vm.latest = !vm.latest}});
      $(this.$refs.failedOnly).checkbox({'onChange': ()=>{vm.failedOnly = !vm.failedOnly}});
  },

  watch: {
    '$route': function(){
      this.fillParams();
      this.fetchList(this.buildQuery());
    }
  },

  methods: {
    fillParams(){
      this.names = this.$route.query.names || '';
      this.ids = this.$route.query.ids || '';
      this.nodes = this.$route.query.nodes || '';
      this.begin = this.$route.query.begin || '';
      this.end = this.$route.query.end || '';
      this.page = this.$route.query.page || 1;
      this.latest = this.$route.query.latest == 'true' ? true : false;
      this.failedOnly = this.$route.query.failedOnly ? true : false;
    },

    fetchList(query){
      this.loading = true;
      var vm = this;
      this.$rest.GET('logs?'+query)
        .onsucceed(200, (resp)=>{
          vm.list = resp.list;
          vm.total = resp.total;
        })
        .onfailed((msg)=>{vm.$bus.$emit('error', msg)})
        .onend(()=>{vm.loading=false})
        .do();
    },

    buildQuery(){
      var params = [];
      if (this.names) params.push('names='+this.names);
      if (this.ids) params.push('ids='+this.ids);
      if (this.nodes) params.push('nodes='+this.nodes);
      if (this.begin) params.push('begin='+this.begin);
      if (this.end) params.push('end='+this.end);
      if (this.failedOnly) params.push('failedOnly=true');
      if (this.page == 0) this.page = 1;
      params.push('page='+this.page);
      if (this.latest) params.push('latest=true');
      return params.join('&');
    },

    submit: function(){
      var query = this.buildQuery()
      var url = '/log?'+query;
      if (this.$route.fullPath == url) {
        this.fetchList(query);
        return;
      }
      this.$router.push(url);
    },

    durationAttention: function(beginTime, endTime){
      var d = new Date(endTime) - new Date(beginTime);
      return d > 3600000*6;
    },

    formatTime: function(log){
      return formatTime(log.beginTime, log.endTime)+'，于 '+log.node+' 耗时 '+formatDuration(log.beginTime, log.endTime);
    },

    showExecuteJobModal: function(jobName, jobGroup, jobId){
      this.$refs.executeJobModal.show(jobName, jobGroup, jobId);
    }
  },
  components: {
    Pager,
    ExecuteJob
  }
}
</script>