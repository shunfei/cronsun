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
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
import JobToolbar from './JobToolbar.vue';
import Dropdown from './basic/Dropdown.vue';

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
    }
  },

  components: {
    JobToolbar,
    Dropdown
  }
}
</script>