define(['Vue', 'text!job/jobEdit.html'], function(Vue, tpl){
  return {
    data: function(){
      return {status: true}
    },
    template: tpl
  };
});