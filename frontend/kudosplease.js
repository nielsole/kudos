  /**
   Kudos Please

   A simple kudos widget without any external lib and it works
   with touch & mouse devices.

   Inspired by dcurt.is

   The heart in this example is served by weloveiconfonts.com,
   but you can just add a value manually (.finish:before{content:''}).

   # 2013 by Tim Pietrusky
   # timpietrusky.com
 **/

 KudosPlease = (function() {

  /*
   * Add <CODE>class</CODE> to <CODE>el</CODE>
   */
  var addClass = function(el, className) {
    if (className) el.classList.add(className);
  },

  /*
   * Remove <CODE>class</CODE> from <CODE>el</CODE>
   */
  removeClass = function(el, className) {
    if (className) el.classList.remove(className);
  },

  /*
   * Returns <CODE>true</CODE> if <CODE>el</CODE> has
   * the <CODE>class</CODE>, <CODE>false</CODE> otherwise
   */
  hasClass = function(el, className) {
    return el.classList.contains(className);
  },

  /*
   * Change the status of the widget and
   * apply 3 different classes for the icon
   * in the middle.
   */
  changeStatus = function(el, state, _$) {
    if (_$.status != undefined) {

      if (el.getAttribute('data-status') != undefined) {
        removeClass(el, _$.status[el.getAttribute('data-status')]);
      }

      addClass(el, _$.status[state]);
      el.setAttribute('data-status', state);
    }
  },

  /*
   * State: finished (kudos given)
   */
  finish = function(el, increase, _$) {
    // Finished
    addClass(el, 'finish');
    changeStatus(el, 'gamma', _$);

    increase = increase || false;
    amount = loadAmount(parseInt(el.getAttribute('data-id'), 10), _$);

    if (increase) {
      ++amount;

      // Update kudos via fetch
      request(el, 'POST', _$);
    }
  },
  /*
   * Enter the element
   */
  enter = function(e, el, _$) {
    var id = -1;

    // Do the kudo twist
    if (!hasClass(el, 'finish')) {
      // Activate the kudo twist
      addClass(el, 'active');

      // Start timeout
      id = setTimeout(function() {
        removeClass(el, 'active');
        finish(el, true, _$);
      }, _$.duration);

      // Add timeout id to global object
      _$.timer[el.getAttribute('data-id')] = id;
    }
  },
   /*
    * Touch enter the element
    */
  touchEnter = function(e, el, _$){
       // Only execute on single finger touch to allow zooming.
       if (e.touches.length === 1){
          // prevent from propagation and preventDefault. So we can use both touch events and mouse events.
          e.stopPropagation();
          e.preventDefault();

          // Execute "normal" enter function
          enter(e, el, _$);
       }
  },

  /*
   * Leave the element
   */
  out = function(e, el, _$) {
    if (!hasClass(el, 'finish')) {
      removeClass(el, 'active');
      clearTimeout(_$.timer[el.getAttribute('data-id')]);
    }
  },
  /*
   * Bind event
   */
  on = function(el, event, func, _$) {
    el.addEventListener(event, function(e) { func(e, el, _$); }, false);
  },

  /*
   * Saves the amount of a specific widget into localStorage
   * when <CODE>persistent</CODE> is <CODE>true</CODE>.
   */
  save = function(el, amount, _$) {
    if (_$.persistent) {
      window.localStorage.setItem('kudos:saved:' + el.getAttribute('data-url'), amount);
    }
  },

  /*
   * Loads the amount of a specific widget from the localStorage
   * when <CODE>persistent</CODE> is <CODE>true</CODE>.
   */
  loadAmount = function(id, _$) {
    var result = _$.elements[id].getAttribute('data-amount') || 0;

    if (_$.persistent) {
      if ((amount = window.localStorage.getItem('kudos:saved:' + _$.elements[id].getAttribute('data-url'))) != undefined) {
        result = amount;
      }
    }

    return result;
  },

  /*
   * Create a fetch request to the backend to keep track of the kudos counter.
   * The backend URL is read from the element's data-backend attribute, falling
   * back to the default hosted instance.
   */
  request = function(el, type, _$) {
    var backendUrl = el.getAttribute('data-backend') || 'https://kudos.niels-ole.com/';
    var url = backendUrl + '?url=' + encodeURIComponent(el.getAttribute('data-url'));

    fetch(url, { method: type })
      .then(function(res) { return res.text(); })
      .then(function(amount) {
        el.setAttribute('data-amount', amount);

        if (type === 'GET') {
          // Set status based on amount
          changeStatus(el, amount == 0 ? 'alpha' : 'beta', _$);

          // When persistence is activated and a value was saved for the URL,
          // the kudos widget is in status gamma (finished)
          if (_$.persistent
           && window.localStorage.getItem('kudos:saved:' + el.getAttribute('data-url')) != null) {
            changeStatus(el, 'gamma', _$);
          }
        }

        if (type === 'POST') {
          save(el, amount, _$);
        }
      });
  };

  /*
   * Constructor
   */
  function KudosPlease(args) {
    // All widgets
    this.elements = document.querySelectorAll(args.el);
    // Set the status
    this.status = args.status;
    // Is localStorage enabled?
    this.persistent = args.persistent != undefined ? (args.persistent && window.localStorage != undefined) : true;
    // Duration of activation
    this.duration = args.duration;
    // setTimeout-ID's
    this.timer = {};

    for (var i = 0; i < this.elements.length; i++) {
      var el = this.elements[i];

      // Delete all elements from localStorage
      // localStorage.setItem('kudos:saved:'+el.getAttribute('data-url'), 0);

      // Identify element
      el.setAttribute('data-id', i);

      // Load kudos via fetch
      request(el, 'GET', this);

      // Amount is 0
      if (loadAmount(i, this) == 0) {
        // Set kudos amount
        el.setAttribute('data-amount', 0);

        // Init timer id
        this.timer[i] = -1;

        // Events, both touch and mouse
        on(el, 'touchstart', touchEnter, this);
        on(el, 'touchend', out, this);
        on(el, 'mouseover', enter, this);
        on(el, 'mouseout', out, this);

      // Load the amount and display it, because user already voted
      } else {
        finish(el, false, this);
      }
    }

    return this;
  }

  return KudosPlease;
 })();
