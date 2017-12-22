<template>
<div class="ui search selection dropdown" v-bind:class="{multiple: multiple}">
  <input type="hidden">
  <div class="default text">{{title}}</div>
  <i class="dropdown icon"></i>
  <div class="menu">
    <div class="item" v-for="item in items" v-bind:data-value="typeof item === 'object' ? item.value : item">{{typeof item === 'object' ? item.name : item}}</div>
  </div>
</div>
</template>

<script>
export default {
  name: 'dropdown',
  props: ['title', 'items', 'allowAdditions', 'multiple', 'selected'],

  data: function(){
    return {}
  },

  mounted: function() {
    if (!this.title || this.title.length === 0) {
      this.title = this.$L(multiple ? 'select groups' : 'select a group');
    }

    var vm = this;
    $(this.$el).dropdown({
      allowAdditions: !!this.allowAdditions,
      hideAdditions: false,
      forceSelection: false,
      onChange: function(value, text, $choice){
        vm.$emit('change', value, text);
      }
    });
    setTimeout(()=>{
      $(vm.$el).dropdown('set exactly', vm.selected).dropdown('refresh');
    }, 500);
  },

  updated: function(){
    $(this.$el).dropdown('set exactly', this.selected);
  }
}
</script>
