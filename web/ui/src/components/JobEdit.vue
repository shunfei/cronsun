<template>
  <form class="ui form segment" v-on:submit.preven>
    <h3 class="ui header">新建任务</h3>
    <div class="two fields">
      <div class="field">
        <label>任务名称</label>
        <input type="text" v-model="name" placeholder="任务名称">
      </div>
       <div class="field">
        <label>任务分组</label>
        <div class="ui dropdown selection">
          <input type="hidden">
          <div class="default text">选择分组</div>
          <i class="dropdown icon"></i>
          <div class="menu">
            <div class="item" data-value="default">Default</div>
            <div class="item" data-value="ssp">SSP</div>
            <div class="item" data-value="dc">数据中心</div>
          </div>
        </div>
      </div>
    </div>
    <div class="field">
      <label>任务脚本</label>
      <input type="text" v-model="cmd" placeholder="任务脚本">
    </div>
    <div class="field">
      <div class="ui toggle checkbox">
        <input type="checkbox" class="hidden" v-bind:checked="!pause">
        <label>{{pause ? '暂停' : '开启'}}</label>
      </div>
    </div>
    <button class="fluid blue ui button" type="submit">创建</button>
  </form>
</template>

<script>
export default {
  name: 'job-edit',
  data: function(){
      return {
        name: '',
        group: 'default',
        cmd: '',
        pause: false
      }
  },

  mounted: function(){
    var vm = this;
    $(this.$el).find('.checkbox').checkbox({
      onChange: function(){
        vm.pause = !vm.pause;
      }
    });

    $(this.$el).find('.dropdown').dropdown({
      allowAdditions: true,
      onChange: function(value, text, $choice){
        vm.group = value;
      }
    }).dropdown('set selected', this.group);
  }
}
</script>