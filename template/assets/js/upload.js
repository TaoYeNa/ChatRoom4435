var res = '';
var reader;
function showImg(obj) {
    var files = obj.files
    // document.getElementById("msger").innerHTML = getImgsByUrl(files)
    getImgsByFileReader(document.getElementById("imgContainer"), files)
}

// window.URL.createObjectURL(file) read file and show picture
function getImgsByUrl(files) {
    var elements = res
    // for (var i = 0; i < files.length; i++) {
    //     var url = window.URL.createObjectURL(files[i])
    //     elements += "< img src='"+ url + "' style='width: 40px; height: 40px; vertical-align: middle; margin-right: 5px;' />"
    // }

    return elements
}

// use FileReader read file and show image
function getImgsByFileReader(el, files) {

    for (var i = 0; i < files.length; i++) {
        if(!/image\/\w+/.test(files[i].type)){
            alert("Please make sure you are uploading img");
            return false;
        }
        let img = document.createElement('img',)
        img.setAttribute('style', 'width: 40px; height: 40px; vertical-align: middle; margin-right: 5px;')
        img.setAttribute('id' , 'image')
        el.appendChild(img)
         reader = new FileReader()
        reader.onload = function(e) {
            img.src = e.target.result
            res = e.target.result
        }
        reader.readAsDataURL(files[i])
    }
}


