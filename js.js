// ==UserScript==
// @name         New Userscript
// @namespace    http://tampermonkey.net/
// @version      0.1
// @description  try to take over the world!
// @author       You
// @match        http*://*/*

// @require      https://cdn.jsdelivr.net/npm/jquery@2.2.4/dist/jquery.min.js
// @require      https://cdn.jsdelivr.net/npm/lovefield@2.1.12/dist/lovefield.min.js
// @require      https://cdn.jsdelivr.net/npm/sweetalert2@9

// @run-at       document-idle
// @grant        GM_xmlhttpRequest
// @grant        GM_addStyle
// @grant        GM_getValue
// @grant        GM_setValue
// @grant        GM_deleteValue
// @grant        GM_notification
// @grant        GM_setClipboard
// @grant        GM_getResourceURL
// @grant        GM_registerMenuCommand
// @grant        GM_info

// @connect     114.taobao.com

// @icon         data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==
// ==/UserScript==

(function() {
    'use strict';

    function javDBScriptPreDownload(){
        if( (/(javbd)*\/v\/*/g).test(document.URL) ) {
            return
        }
        if( !(/(javbd.com)*\/*/g).test(document.URL) ) {
            return
        }
        console.log("javdb.com")
        //$('div.grid-item.column .uid ')[0]
        let javList = $('div.grid-item.column .uid ')

        for (const v of javList) {
            //console.log(v.getInnerHTML());
            let javID = v.getInnerHTML()
            // let $img = $(`<img name="javRealImg" title="无图" class=""></img>`);
            // $img.attr("src", "https://114.taobao.com/download/"+javID);
            // $(jav).after($img);

            GM_xmlhttpRequest( {
                method:     'GET',
                url:        "https://114.taobao.com/download/"+javID,
                onload:     function (responseDetails) {
                    console.log (javID);
                }
            } );
        }
    }

    function javDBScript(){
        // if(document.domain != "www.google.com"){
        //     return
        // }

        if( !(/(javbd)*\/v\/*/g).test(document.URL) ) {
            return
        }

        GM_addStyle(`
                .min {width:66px;min-height: 233px;height:auto;cursor: pointer;} /*Common.addAvImg使用*/
                .col-md-3 {width: 20%;padding-left: 18px; padding-right: 2px;}
                .col-xs-12,.col-md-12 {padding-left: 2px; padding-right: 0px;}
                .col-md-7 {width: 79%;padding-left: 2px;padding-right: 0px;}
                .col-md-9 {width: max-content;}
                .col-md-offset-1 {margin-left: auto;}
                .hobby {display: inline-block;float: left;}
                .hobby_mov {width: 75%;}
                .hobby_p {color: white;font-size: 40px;margin: 0 0 0px;display: inline-block;text-align: right;width: 100%;}
            `);


        //get javID
        let meta = document.getElementsByClassName("title is-4")[0].getElementsByTagName('strong')[0];
        let arr = meta.textContent.split(" ");
        let javID = arr[0];
        console.log("javID:" + javID);


        let divEle = $("div[class='video-meta-panel']")[0];

        let $img = $(`<img name="javRealImg" title="" class=""></img>`);
        $img.attr("src", "https://114.taobao.com/img/"+javID);
        $img.attr("style", "float: left;cursor: pointer;max-width: 100%;");

        if (divEle) {
            $(divEle).after($img);

            $img.click(function () {
                $(this).toggleClass('min');
                if ($(this).attr("class")) {
                    this.parentElement.parentElement.scrollIntoView();
                }
            });
        }

    }

    function mainRun() {
        console.log("start")
        javDBScript();
        javDBScriptPreDownload()
    }
    mainRun();

})();