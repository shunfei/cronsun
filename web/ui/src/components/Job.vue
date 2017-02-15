<style scope>
  .clearfix:after {content:""; clear:both; display:table;}
</style>
<template>
  <div>
    <JobToolbar class="clearfix"/>
    <form class="ui form">
      <div class="field">
        <label>选择一个分组显示其下的任务</label>
        <Dropdown title="选择分组" v-bind:items="groups" v-on:change="changeGroup"/>
      </div>
    </form>
    <table class="ui hover celled striped blue table" v-if="jobs.length > 0">
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
          <td>{{job.name}}</td>
          <td>
            <span v-if="!job.latestStatus">-</span>
            <span v-else>{{formatTime(job.latestStatus.beginTime, job.latestStatus.endTime)}}，耗时 {{formatDuration(job.latestStatus.beginTime, job.latestStatus.endTime)}}</span>
          </td>
          <td>
            <span v-if="!job.latestStatus">-</span>
            <router-link v-else :to="'/log/'+job.latestStatus.refLogId">{{job.latestStatus.success ? '成功' : '失败'}}</router-link>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import JobToolbar from './JobToolbar.vue';
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
    this.$rest.GET('job/groups').onsucceed(200, (resp)=>{
      !resp.includes('default') && resp.unshift('default');
      vm.groups = resp;
    }).do();
  },

  methods: {
    changeGroup: function(val, text){
      var vm = this;
      this.group = val;
      this.refreshList();
    },

    refreshList: function(){
      var vm = this;
      this.$rest.GET('job/group/'+this.group).onsucceed(200, (resp)=>{
        vm.jobs = resp;
        vm.$nextTick(()=>{
          $(vm.$el).find('table .ui.dropdown').dropdown();
        });
      }).do();
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
        vm.refreshList();
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
    JobToolbar,
    Dropdown,
    Pager
  }
}
</script>