<template>
  <div>
    <form class="ui form" v-bind:class="{loading:loading}" v-on:submit.prevent>
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
          <th class="center aligned">状态码</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="log in list">
          <td><router-link class="item" :to="/log/+log.id">{{log.name}}</router-link></td>
          <td>{{log.node}}</td>
          <td :class="{warning: durationAttention(log.beginTime, log.endTime)}"><i class="attention icon" v-if="durationAttention(log.beginTime, log.endTime)"></i> {{formatTime(log.beginTime, log.endTime)}}，{{formatDuration(log.beginTime, log.endTime)}}</td>
          <td :class="{error: log.exitCode != 0}">{{log.exitCode}}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
export default {
  name: 'log',
  data: function(){
    return {
      loading: false,
      names: '',
      nodes: '',
      begin: '',
      end: '',
      list: []
    }
  },

  mounted: function(){
    this.names = this.$route.query.names;
    this.nodes = this.$route.query.nodes;
    this.begin = this.$route.query.begin;
    this.end = this.$route.query.end;

    if (this.names || this.nodes || this.begin || this.end) this.submit();
  },

  methods: {
    submit: function(){
      this.loading = true;
      var vm = this;
      this.$rest.GET('logs?names='+this.name+'&nodes='+this.nodes+'&begin='+this.begin+'&end='+this.end
      )
        .onsucceed(200, (resp)=>{vm.list = resp})
        .onfailed((resp)=>{console.log(resp)})
        .onend(()=>{vm.loading=false})
        .do();
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
      s += this._formatNumber(t.getDate(), 2) + ' ' + this._formatNumber(t.getHours(), 2) + ':' + this._formatNumber(t.getMinutes());
      return s;
    },

    // i > 0
    _formatNumber: function(i, len){
      var n = Math.ceil(Math.log10(i));
      if (n >= len) return i.toString();
      return '0'.repeat(len-n) + i.toString(); 
    }
  }
}
</script>