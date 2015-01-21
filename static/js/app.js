(function(){
    var timelineData;
    var margin = {top:40, right:40, bottom:40, left:40};
    var kWidth = document.body.offsetWidth - margin.left - margin.right;
    var height = 300 - margin.top;
    var downHeight = 200 - margin.bottom;

    var fill = d3.scale.category20(),
        fontSize = d3.scale.sqrt().range([12,100]);

    function addMarker(d, svg, x, y, i){
        var radius = 4,
            xPos = Math.round(x(i)) + 0.5 - radius,
            yPos = y(d.quantity) - radius,
            yZeroPos = height - yPos,
            yKeywordPos = yZeroPos + (i % 5 + 1) * 24;

        var g = svg.append('g').attr('class', 'marker').attr('transform', 'translate(' + xPos + ', ' + yPos + ')');

        g.append('circle')
            .attr('class','marker-bg')
            .attr('cx', radius)
            .attr('cy', radius)
            .attr('r', radius);

        g.append('line')
            .attr('class','line-up')
            .attr('x1',radius)
            .attr('y1',radius)
            .attr('x2',radius)
            .attr('y2',yZeroPos);

        g.append('text')
            .attr('x',radius)
            .attr('y',-10)
            .text(d.quantity);

        g.append('line')
            .attr('class','line-down')
            .attr('x1',radius)
            .attr('y1',yZeroPos)
            .attr('x2',radius)
            .attr('y2',yKeywordPos);

        g.append('text')
            .attr('x',radius)
            .attr('y',yKeywordPos + 14)
            .text(d.keywords[0].value)
            .on('click',function(){addWordCloud(d.keywords)});
    }

    function search(word){
        d3.json('search.json?word=' + word,function(err, data){
            d3.select('#bookList ul').selectAll('li').remove();
            var e = d3.select('#bookList ul').selectAll('li').data(data);
            e.enter().append('li').attr('class','book-item').html(function(d,i){
                var words = '';
                for(var i = 0; i < d.terms.length; i++){
                    words += '<li>' + d.terms[i] + '</li>'
                }
                return '<p><span class="name">' + d.name + '</span><i class="year">' + d.year + '</i><span class="author">' + d.author + '</span></p>' +
                    '<ul class="words">' + words + '</ul>' +
                    '<p class="desc">' + d.desc + '</p>';
            });
        });
    }

    function draw(words) {
        d3.select("#keywords").append("svg")
            .attr('class','word-cloud')
            .attr("width", kWidth + margin.left + margin.right)
            .attr("height", 300)
            .append("g")
            .attr("transform", 'translate(' + (kWidth + margin.left + margin.right)/2 + ',' + '150)')
            .selectAll("text")
            .data(words)
            .enter().append("text")
            .style("font-size", function(d) { return d.size + "px"; })
            .style("font-family", "Impact")
            .style("fill", function(d, i) { return fill(i); })
            .attr("text-anchor", "middle")
            .attr("transform", function(d) {
                return "translate(" + [d.x, d.y] + ")rotate(" + d.rotate + ")";
            })
            .text(function(d) { return d.text; })
            .on('click',function(e){
                d3.select('#bookList form input[name=word]').property('value', e.text);
                search(e.text);
            });
    }

    function addWordCloud(keywords){
        d3.select('#keywords').select('.word-cloud').remove();
        fontSize.domain([keywords[keywords.length-1].count || 1, keywords[0].count])
        d3.layout.cloud().size([kWidth, 300])
        .words(keywords.map(function(d) {
            return {text: d.value, size: d.count};
        }))
        .padding(5)
        .rotate(function() { return ~~(Math.random() * 2) * 90; })
        .font("Impact")
        .fontSize(function(d) { return fontSize(d.size); })
        .on("end", draw)
        .start();
    }

    function drawTimeline(data){
        d3.selectAll('#timeline svg').remove();
        var scale = parseInt(d3.select('#tlScale').property('value'));
        var width = data.length * scale;
        var x = d3.scale.linear().range([0, width]),
            y = d3.scale.linear().range([height,0]);
        var xAxis = d3.svg.axis().scale(x).orient('bottom').innerTickSize(0),
            yAxis = d3.svg.axis().scale(y).orient('left');

        var line = d3.svg.line().interpolate('monotone')
            .x(function(d,i){return x(i);})
            .y(function(d){return y(d.quantity);});

        x.domain([0, data.length-1]);
        y.domain([0, d3.max(data, function(d){return d.quantity;})]).nice();
        xAxis.ticks(data.length).tickFormat(function(d){
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
    }

    d3.json('data.json',function(error, data){
        if(error){
            return console.warn(error);
        }
        timelineData = data;
        drawTimeline(timelineData);
    });

    d3.select('#bookList form').on('submit',function(e){
        d3.event.preventDefault();
        var word = d3.select(this).select('input[name=word]').property('value');
        search(word);
    });

    d3.select('#tlScale').on('change',function(){
        if(timelineData){
            drawTimeline(timelineData);
        }
    });
})();
