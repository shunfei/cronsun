<style scope>
  .show {}
</style>
<template>
  <div class="ui sticky fixed" style="top: 80px; right: 20px; width: 400px;">
    <div v-for="(m, index) in queue" :key="m.id" class="ui floating message transition animate fly left" :class="[m.type, m.animation, m.visiable]">
      <i class="close icon" v-on:click="closeMessage(m.id)"></i>
      <div class="header">{{m.content}}</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'message',
  data(){
    return {
      queue: []
    }
  },
  
  methods: {
    showMessage(type, content){
      var id = Math.random().toString();
      this.queue.push({
        id: id,
        content: content,
        type: type,
        animation: 'in',
        visiable: 'visiable'
      });

      var vm = this;
      setTimeout(()=>{
        vm.closeMessage(id);
      }, 5000);
    },

    closeMessage(id){
      var vm = this;
      for (var i in vm.queue) {
        if (vm.queue[i].id === id) {
          vm.queue[i].animation = 'out';
          setTimeout(()=>{
            for (var i in vm.queue) {
              if (vm.queue[i].id === id) {
                vm.queue.splice(i, 1);
                return;
              }
            }
          }, 600);
          break;
        }
      }
    }
  },
  
  mounted(){
    var vm = this;
    this.$bus.$on('error', (content)=>{
      vm.showMessage('error', content);
    });
    this.$bus.$on('success', (content)=>{
      vm.showMessage('success', content);
    });
    this.$bus.$on('warning', (content)=>{
      vm.showMessage('warning', content);
    });
  }
}
</script>