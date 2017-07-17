var sendXHR = function(opt) {
  var xhr = new XMLHttpRequest();
  xhr.open(opt.method, opt.url, true);

  if (typeof opt.onexception == 'function') {
    var warpExceptionHandler = (msg)=>{
      opt.onexception(msg);
      typeof opt.onend == 'function' && opt.onend(xhr);
    }
    xhr.onabort=()=>{warpExceptionHandler('request aborted.')};
    xhr.onerror=()=>{warpExceptionHandler('request error.')};
    xhr.ontimeout=()=>{warpExceptionHandler('request timeout.')};
  }

  xhr.onreadystatechange = function(){
    if (xhr.readyState !== XMLHttpRequest.DONE) {
      return;
    }

    var data;
    if (typeof xhr.response != 'object') {
      try {
        data = JSON.parse(xhr.response)
      } catch(e) {
        data = xhr.response;
      }
    } else {
      data = xhr.response;
    }

    if (xhr.status != opt.successCode) {
      typeof opt.onfailed == 'function' && opt.onfailed(data, xhr);
    } else if (xhr.status === opt.successCode && typeof opt.onsucceed == 'function') {
      opt.onsucceed(data, xhr);
    } else if (opt.specialHandlers && typeof opt.specialHandlers[xhr.status] === 'function') {
      opt.specialHandlers[xhr.status](data, xhr);
    }

    typeof opt.onend == 'function' && opt.onend(xhr);
  }

  if (typeof opt.data == 'object') {
    opt.data = JSON.stringify(opt.data);
  }
  xhr.send(opt.data);
}

class request {
  constructor(url, method, data, specialHandlers){
    this._url = url;
    this._method = method;
    this._data = data;
    this._specialHandlers = specialHandlers;
  }

  do(){
    sendXHR({
      method: this._method,
      url: this._url,
      data: this._data,
      successCode: this._successCode,
      onsucceed: this._onsucceed,
      onfailed: this._onfailed,
      onexception: this._onexception,
      onend: this._onend,
      specialHandlers: this._specialHandlers
    });
  }

  onsucceed(successCode, f){
    this._successCode = successCode;
    this._onsucceed = f;
    return this;
  }

  onfailed(f){
    this._onfailed = f;
    return this;
  }

  onexception(f){
    this._onexception = f;
    return this;
  }

  onend(f){
    this._onend = f;
    return this;
  }
}

export default class Rest {
  // specialStatusHandle = map[int]function(data, xhr)
  constructor(prefix, defaultFailedHandler, defaultExceptionHandler, specialStatusHandles){
    this.prefix = prefix;
    this.defaultFailedHandler = defaultFailedHandler; // function(url, resp){}
    this.defaultExceptionHandler = defaultExceptionHandler;
    this.specialStatusHandles = specialStatusHandles;
  };

  handleSpecialStatus(code, h) {
    this.mh[code] = h
  };

  GET(url){
    return new request(this.prefix+url, 'GET', null, this.specialStatusHandles)
      .onfailed(this.defaultFailedHandler)
      .onexception(this.defaultExceptionHandler);
  };

  POST(url, formdata){
    return new request(this.prefix+url, 'POST', formdata, this.specialStatusHandles)
      .onfailed(this.defaultFailedHandler)
      .onexception(this.defaultExceptionHandler);
  };

  PUT(url, formdata){
    return new request(this.prefix+url, 'PUT', formdata, this.specialStatusHandles)
      .onfailed(this.defaultFailedHandler)
      .onexception(this.defaultExceptionHandler);
  };

  DELETE(url){
    return new request(this.prefix+url, 'DELETE', null, this.specialStatusHandles)
      .onfailed(this.defaultFailedHandler)
      .onexception(this.defaultExceptionHandler);
  }
}
