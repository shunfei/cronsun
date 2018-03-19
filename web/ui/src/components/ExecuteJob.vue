<template>
  <div class="ui modal">
    <i class="close icon"></i>
    <div class="header">{{$L('executing job: {job}', jobName)}}</div>
    <div class="content">
      <Dropdown :title="$L('node')" :items="nodes" v-on:change="changeNode" style="width:100%"></Dropdown>
    </div>
    <div class="actions">
      <div class="ui deny button">{{$L('cancel')}}</div>
      <div class="ui positive right labeled icon button">{{$L('execute now')}} <i class="checkmark icon"></i></div>
    </div>
  </div>
</template>

<script>
import Dropdown from './basic/Dropdown.vue';

export default {
  name: 'execute-job',
  data(){
    return {
      jobGroup: '',
      jobId: '',
      jobName: '',
      nodes: [],
      selectedNode: '',
      loading: false
    }
  },

  methods: {
    show(jobName, jobGroup, jobId){
      this.jobName = jobName;
      this.jobGroup = jobGroup;
      this.jobId = jobId;
      this.fetchJobNodes();
      $(this.$el).modal({
        closable: false,
        onApprove: this.submit
      }).modal('show');
    },

    hide(){
      $(this.$el).modal('hide');
    },

    fetchJobNodes(){
      var vm = this;
      this.loading = true;
      this.$rest.GET('job/'+this.jobGroup+'-'+this.jobId+'/nodes').
        onsucceed(200, (resp)=>{
          var nodes = [{value: 'all nodes', name: vm.$L('all nodes')}];
          for (var i in resp) {
            nodes.push({value: resp[i], name: vm.$store.getters.hostshows(resp[i])})
          }
          vm.nodes = nodes;
        }).
        onfailed((msg)=>{
          vm.$bus.$emit('error', msg);
          vm.hide();
        }).
        onend(()=>{vm.loading = false}).
        do();
    },

    submit(){
      var vm = this;
      this.loading = true;
      var node = this.selectedNode === 'all nodes' ? '' : this.selectedNode;
      this.$rest.PUT('/job/'+this.jobGroup+'-'+this.jobId+'/execute?node='+node).
        onsucceed(204, ()=>{
          vm.$bus.$emit('success', '执行命令已发送，注意查看任务日志');
          vm.hide();
        }).
        onfailed((msg)=>{vm.$bus.$emit('error', msg)}).
        onend(()=>{vm.loading = false}).
        do();
      return false;
    },

    changeNode(val){
      this.selectedNode = val;
    }
  },

  components: {
    Dropdown
  }
}
</script>
