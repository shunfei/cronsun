import zhCN from './languages/zh-CN';
import en from './languages/en';

import jQuery from 'jquery';
require('jquery.cookie')

var languages = {
  // key is lowercase
  'en': en,
  'zh-cn': zhCN
}

var supported = [
  { name: 'English', code: 'en' },
  { name: '简体中文', code: 'zh-CN' }
]

var language;
var locale = jQuery.cookie('locale') || navigator.language || 'en';
locale = locale.indexOf('en') === -1 ? 'zh-CN' : 'en';
setLocale(locale);

function setLocale(loc) {
  var loc2 = loc.toLowerCase()
  if (languages[loc2]) {
    locale = loc
    language = languages[loc2]
  }
}

function getLocale() {
  return locale
}

function L(k) {
  var t = language[k];
  if (typeof t === 'undefined') {
    t = k;
  }

  var tr = '';
  var inTag = false;
  var num = '';
  var stash = '';
  for (var i = 0; i < t.length; i++) {
    var cc = t[i].charCodeAt();
    if (cc === 123) { // {
      if (inTag) {
        tr += stash;
        stash = '';
      }
      inTag = true;
    } else if (inTag && cc >= 48 && cc <= 57) {
      num += t[i];
    } else if (inTag && cc === 125) { // }
      var index = parseInt(num);
      if (arguments.length - 1 > index) {
        tr += arguments[index + 1];
      }
      num = '';
      stash = '';
      inTag = false;
      continue;
    } else {
      inTag = false;
      tr += stash;
      stash = '';
    }

    if (inTag) {
      stash += t[i];
    } else {
      tr += t[i];
    }
  }

  return tr;
}

const lang = { getLocale, setLocale, supported, L };
export default lang;
