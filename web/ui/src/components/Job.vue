<style scope>
  .clearfix:after {content:""; clear:both; display:table;}
</style>
<template>
  <div>
    <div class="clearfix">
      <router-link class="ui right floated primary button" to="/job/create"><i class="add to calendar icon"></i> 新任务</router-link>
      <button class="ui right floated icon button" v-on:click="refresh"><i class="refresh icon"></i></button>
    </div>
    <form class="ui form">
      <div class="field">
        <label>选择一个分组显示其下的任务</label>
        <Dropdown title="选择分组" v-bind:items="groups" v-on:change="changeGroup" selected="group"/>
      </div>
    </form>
    <table class="ui hover blue table" v-if="jobs.length > 0">
      <thead>
        <tr>
          <th class="collapsing center aligned">操作</th>
          <th class="collapsing center aligned">状态</th>
          <th width="200px" class="center aligned">分组</th>
          <th class="center aligned">名称</th>
          <th class="center aligned">最近执行时间</th>
          <th class="center aligned">执行结果</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(job, index) in jobs">
          <td class="center aligned">
            <div class="ui icon dropdown">
              <i class="content icon"></i>
              <div class="menu">
                <div class="item" v-on:click="$router.push('/job/edit/'+job.group+'/'+job.id)">编辑</div>
                <div class="item" v-if="job.pause" v-on:click="changeStatus(job.group, job.id, index, !job.pause)">开启</div>
                <div class="item" v-if="!job.pause" v-on:click="changeStatus(job.group, job.id, index, !job.pause)">暂停</div>
                 <div class="divider"></div>
                <div class="item" style="color:red;" v-on:click="removeJob(job.group, job.id, index)">删除</div>
              </div>
            </div>
          </td>
          <td class="center aligned"><i class="icon" v-bind:class="{pause: job.pause, play: !job.pause, green: !job.pause}"></i></td>
          <td>{{job.group}}</td>
          <td><router-link :to="'/job/edit/'+job.group+'/'+job.id">{{job.name}}</router-link></td>
          <td>
            <span v-if="!job.latestStatus">-</span>
            <span v-else>{{formatTime(job.latestStatus.beginTime, job.latestStatus.endTime)}}，于 {{job.latestStatus.node}} 耗时 {{formatDuration(job.latestStatus.beginTime, job.latestStatus.endTime)}}</span>
          </td>
          <td :class="{error: job.latestStatus && !job.latestStatus.success}">
            <span v-if="!job.latestStatus">-</span>
            <router-link v-else :to="'/log/'+job.latestStatus.refLogId">{{job.latestStatus.success ? '成功' : '失败'}}</router-link> |
            <router-link :to="{path: 'log', query: {latest:true, ids: job.id}}">latest</router-link>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';
import Pager from './basic/Pager.vue';

export default {
  name: 'job',
  data: function(){
    return {
      groups: [],
      group: '',
      jobs: []
    }
  },
  
  mounted: function(){
    var vm = this;
    this.group = this.$route.query.group || '';

    this.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      resp.unshift({value: '', name: '所有任务'});
      vm.groups = resp;
      this.fetchList(this.buildQuery());
    }).do();
  },

  watch: {
    '$route': function(){
      this.group = this.$route.query.group || '';
      this.fetchList(this.buildQuery());
    }
  },

  methods: {
    changeGroup: function(val, text){
      var vm = this;
      this.group = val;
      this.$router.push('job?'+this.buildQuery());
    },

    buildQuery: function(){
      var params = [];
      if (this.group) params.push('group='+this.group);
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

      if (s.length == 0) s = "0 毫秒";
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
    Dropdown,
    Pager
  }
}
</script>