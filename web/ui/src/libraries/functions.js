var formatDuration = function(beginTime, endTime){
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
  if (d >= 1) s += d.toString() + ' 毫秒';

  if (s.length == 0) s = "0 毫秒";
  return s;
}

var formatTime = function(beginTime, endTime){
  var now = new Date();
  var bt = new Date(beginTime);
  var et = new Date(endTime);
  var s = _formatTime(now, bt) + ' ~ ' + _formatTime(now, et);
  return s;
}

var _formatTime = function(now, t){
  var s = '';
  if (now.getFullYear() != t.getFullYear()) {
    s += t.getFullYear().toString() + '-';
  }
  s += formatNumber(t.getMonth()+1, 2).toString() + '-';
  s += formatNumber(t.getDate(), 2) + ' ' + formatNumber(t.getHours(), 2) + ':' + formatNumber(t.getMinutes(), 2) + ':' + formatNumber(t.getSeconds(), 2);
  return s;
}

// i > 0
var formatNumber = function(i, len){
  var n = i == 0 ? 1 : Math.ceil(Math.log10(i+1));
  if (n >= len) return i.toString();
  return '0'.repeat(len-n) + i.toString(); 
}

export {formatDuration, formatTime, formatNumber};