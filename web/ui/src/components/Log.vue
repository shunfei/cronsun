<template>
  <div>
    <form class="ui form" method="GET" v-bind:class="{loading:loading}" v-on:submit.prevent>
      <div class="field">
        <label>任务名称</label>
        <input type="text" ref="name" v-model="names"  placeholder="多个名称用英文逗号分隔">
      </div>
      <div class="field">
        <label>运行节点</label>
        <input type="text" ref="name" v-model="nodes" placeholder="ip，多个 ip 用英文逗号分隔">
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
      <div class="field">
        <button class="fluid ui button" type="button" v-on:click="submit">查询</button>
      </div>
    </form>
    <table class="ui selectable green table" v-if="list && list.length > 0">
      <thead>
        <tr>
          <th class="center aligned">任务名称</th>
          <th class="center aligned">运行节点</th>
          <th class="center aligned">执行时间</th>
          <th class="center aligned">运行结果</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="log in list">
          <td><router-link class="item" :to="'/job/edit/'+log.jobGroup+'/'+log.jobId">{{log.name}}</router-link></td>
          <td>{{log.node}}</td>
          <td :class="{warning: durationAttention(log.beginTime, log.endTime)}"><i class="attention icon" v-if="durationAttention(log.beginTime, log.endTime)"></i> {{formatTime(log.beginTime, log.endTime)}}，{{formatDuration(log.beginTime, log.endTime)}}</td>
          <td :class="{error: !log.success}">
            <router-link :to="'/log/'+log.id">{{log.success ? '成功' : '失败'}}</router-link>
          </td>
        </tr>
      </tbody>
    </table>
    <Pager v-if="list && list.length>0" :total="total" :length="5"/>
  </div>
</template>

<script>
import Pager from './basic/Pager.vue';

export default {
  name: 'log',
  data: function(){
    return {
      loading: false,
      names: '',
      nodes: '',
      begin: '',
      end: '',
      list: [],
      total: 0,
      page: 1
    }
  },

  mounted: function(to, from, next){
      this.names = this.$route.query.names || '';
      this.nodes = this.$route.query.nodes || '';
      this.begin = this.$route.query.begin || '';
      this.end = this.$route.query.end || '';
      this.page = this.$route.query.page || 1;
      this.fetchList(this.buildQuery());
  },

  watch: {
    '$route': function(){
      this.names = this.$route.query.names || '';
      this.nodes = this.$route.query.nodes || '';
      this.begin = this.$route.query.begin || '';
      this.end = this.$route.query.end || '';
      this.page = this.$route.query.page || 1;
      
      this.fetchList(this.buildQuery());
    }
  },

  methods: {
    fetchList(query){
      this.loading = true;
      var vm = this;
      this.$rest.GET('/logs?'+query)
        .onsucceed(200, (resp)=>{
          vm.list = resp.list;
          vm.total = resp.total;
        })
        .onfailed((resp)=>{console.log(resp)})
        .onend(()=>{vm.loading=false})
        .do();
    },

    buildQuery(){
      var params = [];
      if (this.names) params.push('names='+this.names);
      if (this.nodes) params.push('nodes='+this.nodes);
      if (this.begin) params.push('begin='+this.begin);
      if (this.end) params.push('end='+this.end);
      if (this.page == 0) this.page = 1;
      params.push('page='+this.page);
      return params.join('&');
    },

    submit: function(){
      this.$router.push('/log?'+this.buildQuery());
    },

    durationAttention: function(beginTime, endTime){
      var d = new Date(endTime) - new Date(beginTime);
      return d > 3600000*6;
    },

    formatDuration: function(beginTime, endTime){
      var d = new Date(endTime) - new Date(beginTime);
      var s = '';
      var day = d/86400000;
      if (day >= 1) s +=  day.toString() + ' 天 '; 
      
      d = d%86400000;
      var hour = d/3600000;
      if (hour >= 1) s += hour.toString() + ' 小时 ';

      d = d%3600000;
      var min = d/60000;
      if (min >= 1) s += min.toString() + ' 分钟 ';

      d = d%60000;
      var sec = d/1000;
      if (sec >= 1) s += sec.toString() + ' 秒 ';

      d = d%1000;
      if (d >= 1) s = d.toString() + ' 毫秒';

      return s;
    },

    formatTime: function(beginTime, endTime){
      var now = new Date();
      var bt = new Date(beginTime);
      var et = new Date(endTime);
      var s = this._formatTime(now, bt) + ' ~ ' + this._formatTime(now, et);
      return s;
    },

    _formatTime: function(now, t){
      var s = '';
      if (now.getFullYear() != t.getFullYear()) {
        s += t.getFullYear().toString() + '-';
      }
      s += this._formatNumber(t.getMonth()+1, 2).toString() + '-';
      s += this._formatNumber(t.getDate(), 2) + ' ' + this._formatNumber(t.getHours(), 2) + ':' + this._formatNumber(t.getMinutes(), 2) + ':' + this._formatNumber(t.getSeconds(), 2);
      return s;
    },

    // i > 0
    _formatNumber: function(i, len){
      var n = i == 0 ? 1 : Math.ceil(Math.log10(i+1));
      if (n >= len) return i.toString();
      return '0'.repeat(len-n) + i.toString(); 
    }
  },
  components: {
    Pager
  }
}
</script>