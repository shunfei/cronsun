<template>
  <div>
    <form class="ui form" method="GET" v-bind:class="{loading:loading}" v-on:submit.prevent v-on:keyup.enter="submit">
      <div class="two fields">
        <div class="field">
          <label>{{$L('job name')}}</label>
          <input type="text" v-model="names" :placeholder="$L('multiple names can separated by commas')">
        </div>
        <div class="field">
          <label>{{$L('job ID')}}</label>
          <input type="text" v-model="ids" :placeholder="$L('multiple IDs can separated by commas')">
        </div>
      </div>
      <div class="field">
        <label>{{$L('node')}}</label>
        <input v-if="$store.getters.showWithHostname" type="text" v-model="hostnames" :placeholder="$L('multiple Hostnames can separated by commas')">
        <input v-else type="text" v-model="ips" :placeholder="$L('multiple IPs can separated by commas')">
      </div>
      <div class="two fields">
        <div class="field">
          <label>{{$L('starting date')}}</label>
          <input type="date" v-model="begin">
        </div>
        <div class="field">
          <label>{{$L('end date')}}</label>
          <input type="date" v-model="end">
        </div>
      </div>
      <div class="two fields">
        <div class="filed">
          <div ref="latest" class="ui checkbox">
            <input type="checkbox" class="hidden" v-model="latest">
            <label>{{$L('latest result of each job on each node')}}</label>
          </div>
        <div class="filed">
          <div ref="failedOnly" class="ui checkbox">
            <input type="checkbox" class="hidden" v-model="failedOnly">
            <label>{{$L('failure only')}}</label>
          </div>
        </div>
        </div>
      </div>
      <div class="field">
        <button class="fluid ui button" type="button" v-on:click="submit">{{$L('submit query')}}</button>
      </div>
    </form>
    <table class="ui selectable green table" v-if="list && list.length > 0">
      <thead>
        <tr>
          <th class="center aligned">{{$L('job name')}}</th>
          <th class="center aligned">{{$L('executing node')}}</th>
          <th class="center aligned">{{$L('executing user')}}</th>
          <th class="center aligned">{{$L('executing time')}}</th>
          <th class="center aligned">{{$L('executing result')}}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="log in list">
          <td><router-link class="item" :to="'/job/edit/'+log.jobGroup+'/'+log.jobId">{{log.name}}</router-link></td>
          <td :title="log.node">{{$store.getters.hostshows(log.node)}}</td>
          <td>{{log.user}}</td>
          <td :class="{warning: durationAttention(log.beginTime, log.endTime)}"><i class="attention icon" v-if="durationAttention(log.beginTime, log.endTime)"></i> {{formatTime(log)}}</td>
          <td :class="{error: !log.success}">
            <router-link :to="'/log/'+log.id">{{$L(log.success ? 'successed' : 'failed')}}</router-link> |
            <a href="#" :title="$L('click to select a node and re-execute job')" v-on:click.prevent="showExecuteJobModal(log.name, log.jobGroup, log.jobId)"><i class="icon repeat"></i></a>
          </td>
        </tr>
      </tbody>
    </table>
    <Pager v-if="list && list.length>0" :total="total" :maxBtn="5"/>
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
      hostnames: '',
      ips: '',
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
      this.hostnames = this.$route.query.hostnames || '';
      this.ips = this.$route.query.ips || '';
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
      if (!this.$store.getters.showWithHostname && this.ids) params.push('ids='+this.ids);
      if (this.$store.getters.showWithHostname && this.hostnames) params.push('hostnames='+this.hostnames);
      if (this.ips) params.push('ips='+this.ips);
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
      return this.$L('took {times}, {begin ~ end}', formatDuration(log.beginTime, log.endTime), formatTime(log.beginTime, log.endTime));
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
