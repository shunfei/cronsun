<template>
  <div>
    <JobToolbar/>
    <form class="ui form">
      <div class="field">
        <label>任务分组</label>
        <Dropdown title="选择分组" v-bind:items="groups" v-on:change="changeGroup"/>
      </div>
    </form>
  </div>
</template>

<script>
import JobToolbar from './JobToolbar.vue';
import Dropdown from './basic/Dropdown.vue';

export default {
  name: 'job',
  data: function(){
    return {
      groups: []
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
    changeGroup: function(){}
  },

  components: {
    JobToolbar,
    Dropdown
  }
}
</script>