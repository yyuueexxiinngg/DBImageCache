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

// @icon         data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    function javDBScript(){
        if(document.domain != "www.google.com"){
            return
        }
        // if( !(/(javbd)*\/v\/*/g).test(document.URL) ) {
        //     return
        // }
        console.log("123")


        //get javID
        // let meta = document.getElementsByClassName("title is-4")[0].getElementsByTagName('strong')[0];
        // let arr = meta.textContent.split(" ");
        // let javID = arr[0];
        //console.log("javID:" + javID);



        //let divEle = $("div[class='video-meta-panel")[0];
        let divEle = $("div[class='RNNXgb")[0];
        //let url = "<img src=\"http://127.0.0.1:8080/static/image.png\">";
        let img = document.createElement("img");//创建一个标签
        img.setAttribute("src", "https://127.0.0.1:8080/static/image.png");
        //img.setAttribute("src", "https://upload.jianshu.io/users/upload_avatars/5957/6d510770-b55c-465f-80bb-10aab2714cfa.jpg?imageMogr2/auto-orient/strip|imageView2/1/w/96/h/96/format/webp");
        //$(divEle).attr("id", "video_info");
        if (divEle) {
            $(divEle).after(img);
            //如果存在min就去除min,否则不存在则添加上min
            $img.click(function () {
                $(this).toggleClass('min');
                if ($(this).attr("class")) {
                    this.parentElement.parentElement.scrollIntoView();
                }
            });
        }

    }

    function mainRun() {
        console.log("123")
        javDBScript();
    }
    mainRun();

})();