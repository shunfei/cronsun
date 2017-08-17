<template>
  <div v-if="confLoading" class="ui segment" style="height:180px; z-index: 10;">
    <div class="ui active inverted dimmer">
      <div class="ui large text loader">{{$L('loading configurations')}}...</div>
    </div>
  </div>
  <JobEditForm v-else/>
</template>

<script>
import JobEditForm from './JobEditForm.vue';

export default {
  name: 'job-edit',
  data: function(){
      return {
        confLoading: true
      }
  },

  mounted: function(){
    var vm = this;
    this.$onConfigLoaded(() => vm.confLoading = false);
  },

  components: {
    JobEditForm
  }
}
</script>
