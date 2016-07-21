(function(){
    var defPage = 1;
    var defSize = 50;

    var timelineData;
    var margin = {top:40, right:40, bottom:40, left:40};
    var kWidth = 400, kHeight = 400;
    var height = 300 - margin.top;
    var downHeight = 200 - margin.bottom;

    var fill = d3.scale.category20(),
        fontSize = d3.scale.sqrt().range([12,60]);

    var kcDrawed = false, kSearched = false;
    var words = [];
    var layout = d3.layout.cloud()
        .timeInterval(10)
        .size([kWidth,kHeight])
        .spiral('archimedean')//archimedean,rectangular
        .padding(5)
        .rotate(function() { return ~~(Math.random() * 2) * 90; })
        .font("Impact")
        .fontSize(function(d) { return fontSize(d.size); })
        .on("end", draw);

    var kSvg = d3.select("#keywords").append("svg")
            .attr('class','word-cloud')
            .attr("width", kWidth)
            .attr("height", kHeight);
    var kBg = kSvg.append("g")
            .attr("transform", 'translate(' + kWidth/2 + ',' + kHeight/2 + ')');

    var colorScale = d3.scale.linear().range([50,0]);
    var years = [];

    function calColor(quantity){
        return 'hsl(' + colorScale(quantity) + ',100%,50%)';
    }

    function addMarker(d, svg, x, y, i){
        var radius = 4,
            xPos = Math.round(x(i)) + 0.5 - radius,
            yPos = y(d.quantity) - radius,
            yZeroPos = height - yPos,
            yKeywordPos = yZeroPos + (i % 5 + 1) * 24;

        var g = svg.append('g').attr('class', 'marker').attr('transform', 'translate(' + xPos + ', ' + yPos + ')');
        var fc = calColor(d.quantity);

        g.append('line')
            .attr('class','line-up')
            .attr('x1',radius)
            .attr('y1',radius)
            .attr('x2',radius)
            .attr('y2',yZeroPos);

        g.append('text')
            .attr('class','tip')
            .attr('x',radius)
            .attr('y',-10)
            .attr('fill',fc)
            .text(d.quantity);

        g.append('line')
            .attr('class','line-down')
            .attr('x1',radius)
            .attr('y1',yZeroPos)
            .attr('x2',radius)
            .attr('y2',yKeywordPos);

        g.append('text')
            .attr('class','word')
            .attr('x',radius)
            .attr('y',yKeywordPos + 14)
            .text(d.keywords[0].value)
            .on('click',function(){
                addWordCloud(d);
                search(d3.select(this).text(), d.year, defPage, defSize);
            });

        g.append('circle')
            .attr('class','marker-bg')
            .attr('cx', radius)
            .attr('cy', radius)
            .attr('r', radius)
            .attr('fill',fc);


    }

    function search(word,year,page,limit){
        d3.select('#bookList form input[name=word]').property('value', word);
        d3.select('#bookList form select[name=year]').property('value', year);
        var start = (page - 1) * limit;
        d3.json('search.json?word=' + word + '&year=' + year + '&start=' + start + '&limit=' + limit, function(err, data){
            d3.select('#bookList ul.data-list').selectAll('li').remove();
            var e = d3.select('#bookList ul.data-list').selectAll('li').data(data.docs);
            e.enter().append('li').attr('class','book-item').html(function(d,i){
                var words = '';
                for(var i = 0; i < d.terms.length; i++){
                    words += '<li>' + d.terms[i] + '</li>'
                }
                return '<p><a target="_blank" href="' + (d.url ? d.url : '#') + '"><span class="name">' + d.name + '</span></a><i class="year">' + d.year + '</i><span class="author">' + (d.author ? d.author.join(',') : '') + '</span></p>' +
                    '<ul class="words">' + words + '</ul>' +
                    '<p class="desc">' + d.desc + '</p>';
            });
            d3.selectAll('#bookList .book-item li').on('click',function(d){
                var text = d3.select(this).text();
                search(text, '', defPage, limit);
            });
            resetPager(page, limit, data.total);
        });
    }

    function searchPage(page, size){
        var word = d3.select('#bookList form').select('input[name=word]').property('value')||'';
        var year = d3.select('#bookList form').select('select[name=year]').property('value')||'';
        search(word, year, page, size);
    }

    function searchWord(word, page, size){
        var year = d3.select('#bookList form').select('select[name=year]').property('value')||'';
        search(word, year, page, size);
    }

    function resetPager(page, size, total){
        var pagination = d3.select('nav .pagination');
        pagination.selectAll('li').remove();
        if(size >= total){
            pagination.classed('hidden', true);
        }else{
            pagination.classed('hidden', false);
            var tp = Math.ceil(total/size);
            var pd = [];
            for(var i = 1; i <= tp; i++){
                pd.push(i)
            }
            var e = pagination.selectAll('li')
                .data(pd)
                .enter()
                .append('li').attr('class', function(d){
                    return d == page ? 'active' : '';
                }).html(function(d){
                    return '<span>' + d + '</span>';
                }).on('click',function(d){
                    var li = d3.select(this);
                    if(!li.classed('active')){
                        searchPage(d, size);
                    }
                });
        }
    }

    function draw(data) {
        words = data;
        var text = kBg.selectAll("text")
            .data(words);
        text.enter().append("text")
            .attr("transform", function(d) { return "translate(" + [d.x, d.y] + ")rotate(" + d.rotate + ")"; })
            .style("font-size", function(d) { return d.size + "px"; })
            .style("font-family", "Impact")
            .style("fill", function(d, i) { return fill(i); })
            .attr("text-anchor", "middle")
            .attr("transform", function(d) {
                return "translate(" + [d.x, d.y] + ")rotate(" + d.rotate + ")";
            })
            .style("opacity", 1e-6)
            .on('click',function(e){
                searchWord(e.text, defPage, defSize);
            })
            .transition()
            .duration(1000)
            .style("opacity", 1)
            .text(function(d) { return d.text; });
        text.exit().remove();
    }

    function addWordCloud(data){
        var keywords = data.keywords;
        fontSize.domain([keywords[keywords.length-1].count || 1, keywords[0].count])
        words = [];
        layout.words(words).start();
        layout.stop()
            .words(keywords.map(function(d) {
                return {text: d.value, size: d.count};
            }))
            .start();
        kcDrawed = true;
        if(!kSearched && keywords.length > 0){
            search(keywords[0].value, data.year, defPage, defSize);
            kSearched = true;
        }
    }

    function autoTicks(scale,max){
        if(scale < 30){
            return Math.round(max/2);
        }
        return max;
    }

    function drawTimeline(data){
        d3.selectAll('#timeline svg').remove();
        var scale = parseInt(d3.select('#tlScale').property('value'));
        var width = data.length * scale;
        var x = d3.scale.linear().range([0, width]),
            y = d3.scale.linear().range([height,0]);
        var xAxis = d3.svg.axis().scale(x).orient('bottom').innerTickSize(0).tickPadding(5),
            yAxis = d3.svg.axis().scale(y).orient('left');

        var line = d3.svg.line().interpolate('monotone')
            .x(function(d,i){return x(i);})
            .y(function(d){return y(d.quantity);});

        x.domain([0, data.length-1]);
        y.domain([0, d3.max(data, function(d){return d.quantity;})]).nice();
        xAxis.ticks(autoTicks(scale, data.length)).tickFormat(function(d){
            return data[d].year;
        });
        var svg = d3.select('#timeline .wrapper').append('svg')
            .attr('width', width + margin.left + margin.right)
            .attr('height', height + downHeight + margin.top + margin.bottom)
            .append('g')
            .attr('transform', 'translate(' + margin.left + ',' + margin.top + ')');

        //add the x-axis
        svg.append('g')
            .attr('class', 'x axis')
            .attr('transform','translate(0,' + height + ')')
            .call(xAxis);
        //add the y-axis
        /*svg.append('g')
            .attr('class', 'y axis')
            .call(yAxis);*/
        //add the line
        svg.append('path')
            .attr('class', 'line')
            .attr('d', line(data));

        data.forEach(function(d, i){
            addMarker(d, svg, x, y, i);
        });
        if(!kcDrawed){
            addWordCloud(data[0])
        }
    }

    d3.json('data.json',function(error, data){
        if(error){
            return console.warn(error);
        }
        timelineData = data;
        colorScale.domain(d3.extent(data,function(d){return d.quantity;}))
        years = data.map(function(d){return d.year;});
        years.unshift("");
        var opts = d3.select('.search select[name=year]').selectAll('option').data(years);
        opts.enter()
            .append('option')
            .attr('value',function(d){return d;})
            .text(function(d){return d;});
        drawTimeline(timelineData);
    });

    d3.select('#bookList form').on('submit',function(e){
        d3.event.preventDefault();
        var word = d3.select(this).select('input[name=word]').property('value')||'';
        var year = d3.select(this).select('select[name=year]').property('value')||'';
        search(word, year, defPage, defSize);
    });

    d3.select('#tlScale').on('change',function(){
        if(timelineData){
            drawTimeline(timelineData);
        }
    });
})();
