var formatDuration = function(beginTime, endTime){
  var d = new Date(endTime) - new Date(beginTime);
  var s = '';
  var day = Math.floor(d/86400000);
  if (day >= 1) s +=  day.toString() + ' d '; 
  
  d = d%86400000;
  var hour = Math.floor(d/3600000);
  if (hour >= 1) s += hour.toString() + ' hr ';

  d = d%3600000;
  var min = Math.floor(d/60000);
  if (min >= 1) s += min.toString() + ' min ';

  d = d%60000;
  var sec = Math.floor(d/1000);
  if (sec >= 1) s += sec.toString() + ' s ';

  d = Math.floor(d%1000);
  if (d >= 1) s += d.toString() + ' ms';

  if (s.length == 0) s = '0 ms';
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

var split = function(str, sep){
  if (typeof str != 'string' || str.length === 0) return [];
  return str.split(sep || ',');
}

export {formatDuration, formatTime, formatNumber, split};
