package poc_css_chess_scandinavian

func Render(path string) string {
	return `This uses just HTML and CSS animation.

<h4>Scandinavian opening</h4>

` + css() + html()
}

func css() string {
	return `<style type="text/css">
    .board { 
        display: flex;
        flex-wrap: wrap;
        font-size: 30px; 
        width: 8em;
        height: 8em;
        position: relative;
    }
    .floating {
        float: left;
        margin-right: 10px;
    }
    .square { 
        width: 1em; 
        height: 1em; 
        line-height: 1em;
        display: inline-block; 
        margin: 0; 
        vertical-align: middle;
    }
    span.dark { 
        background-color: #8a6; 
        background-color: #a9a; 
    }
    span.light { 
        background-color: #6aa; 
        background-color: #acc; 
        background-color: #ccc;
    }
    .piece { 
        position: absolute; 
        display: inline-block;
        line-height: 1em;
        height: 1em;
        width: 1em;
        padding: 0;
        margin-top: 0.031em;
        text-align: center;
        text-shadow: 1px 1px 1px aliceblue;
        animation-duration: 1s;
        animation-fill-mode: forwards;
    }
    .white.king::before { content: '\2654'; }
    .white.queen::before { content: '\2655'; }
    .white.rook::before { content: '\2656'; }
    .white.bishop::before { content: '\2657'; }
    .white.knight::before { content: '\2658'; }
    .white.pawn::before { content: '\2659'; }
    .black.king::before { content: '\265A'; }
    .black.queen::before { content: '\265B'; }
    .black.rook::before { content: '\265C'; }
    .black.bishop::before { content: '\265D'; }
    .black.knight::before { content: '\265E'; }
    .black.pawn::before { content: '\265F'; }
    hr.clear { clear: both; visibility: hidden; }
    .piece.a1 { left: 0em; bottom: 0em; }
    .piece.b1 { left: 1em; bottom: 0em; }
    .piece.c1 { left: 2em; bottom: 0em; }
    .piece.d1 { left: 3em; bottom: 0em; }
    .piece.e1 { left: 4em; bottom: 0em; }
    .piece.f1 { left: 5em; bottom: 0em; }
    .piece.g1 { left: 6em; bottom: 0em; }
    .piece.h1 { left: 7em; bottom: 0em; }
    .piece.a2 { left: 0em; bottom: 1em; }
    .piece.b2 { left: 1em; bottom: 1em; }
    .piece.c2 { left: 2em; bottom: 1em; }
    .piece.d2 { left: 3em; bottom: 1em; }
    .piece.e2 { left: 4em; bottom: 1em; animation-delay: 1s; animation-name: e2e4; }
    @keyframes e2e4 { to { left: 4em; bottom: 3em; } }
    .piece.f2 { left: 5em; bottom: 1em; }
    .piece.g2 { left: 6em; bottom: 1em; }
    .piece.h2 { left: 7em; bottom: 1em; }
    .piece.a3 { left: 0em; bottom: 2em; }
    .piece.b3 { left: 1em; bottom: 2em; }
    .piece.c3 { left: 2em; bottom: 2em; }
    .piece.d3 { left: 3em; bottom: 2em; }
    .piece.e3 { left: 4em; bottom: 2em; }
    .piece.f3 { left: 5em; bottom: 2em; }
    .piece.g3 { left: 6em; bottom: 2em; }
    .piece.h3 { left: 7em; bottom: 2em; }
    .piece.a4 { left: 0em; bottom: 3em; }
    .piece.b4 { left: 1em; bottom: 3em; }
    .piece.c4 { left: 2em; bottom: 3em; }
    .piece.d4 { left: 3em; bottom: 3em; }
    .piece.e4 { left: 4em; bottom: 3em; }
    .piece.f4 { left: 5em; bottom: 3em; }
    .piece.g4 { left: 6em; bottom: 3em; }
    .piece.h4 { left: 7em; bottom: 3em; }
    .piece.a5 { left: 0em; bottom: 4em; }
    .piece.b5 { left: 1em; bottom: 4em; }
    .piece.c5 { left: 2em; bottom: 4em; }
    .piece.d5 { left: 3em; bottom: 4em; }
    .piece.e5 { left: 4em; bottom: 4em; }
    .piece.f5 { left: 5em; bottom: 4em; }
    .piece.g5 { left: 6em; bottom: 4em; }
    .piece.h5 { left: 7em; bottom: 4em; }
    .piece.a6 { left: 0em; bottom: 5em; }
    .piece.b6 { left: 1em; bottom: 5em; }
    .piece.c6 { left: 2em; bottom: 5em; }
    .piece.d6 { left: 3em; bottom: 5em; }
    .piece.e6 { left: 4em; bottom: 5em; }
    .piece.f6 { left: 5em; bottom: 5em; }
    .piece.g6 { left: 6em; bottom: 5em; }
    .piece.h6 { left: 7em; bottom: 5em; }
    .piece.a7 { left: 0em; bottom: 6em; }
    .piece.b7 { left: 1em; bottom: 6em; }
    .piece.c7 { left: 2em; bottom: 6em; }
    .piece.d7 { left: 3em; bottom: 6em; animation-delay: 3s; animation-name: d7d5; }
    @keyframes d7d5 { to { left: 3em; bottom: 4em; } }
    .piece.e7 { left: 4em; bottom: 6em; }
    .piece.f7 { left: 5em; bottom: 6em; }
    .piece.g7 { left: 6em; bottom: 6em; }
    .piece.h7 { left: 7em; bottom: 6em; }
    .piece.a8 { left: 0em; bottom: 7em; }
    .piece.b8 { left: 1em; bottom: 7em; }
    .piece.c8 { left: 2em; bottom: 7em; }
    .piece.d8 { left: 3em; bottom: 7em; }
    .piece.e8 { left: 4em; bottom: 7em; }
    .piece.f8 { left: 5em; bottom: 7em; }
    .piece.g8 { left: 6em; bottom: 7em; }
    .piece.h8 { left: 7em; bottom: 7em; }
    
    .explanations { 
        animation-duration: 2s;
        animation-delay: 4s;
        animation-name: showExplanations;
        animation-fill-mode: forwards;
        animation-easing: ease-in;
        opacity: 0;
    }
    @keyframes showExplanations {
        to { opacity: 1; } 
    }
</style>

`
}

func html() string {
	return `
&nbsp;
<span class="board floating"> 
    <span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span>
	<span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span>
	<span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span>
	<span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span>
	<span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span>
	<span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span>
	<span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span>
	<span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span><span class="dark square"></span><span class="light square"></span>
    <span class="piece a1 white rook"></span>
    <span class="piece b1 white knight"></span>
    <span class="piece c1 white bishop"></span>
    <span class="piece d1 white queen"></span>
    <span class="piece e1 white king"></span>
    <span class="piece f1 white bishop"></span>
    <span class="piece g1 white knight"></span>
    <span class="piece h1 white rook"></span>
    <span class="piece a2 white pawn"></span>
    <span class="piece b2 white pawn"></span>
    <span class="piece c2 white pawn"></span>
    <span class="piece d2 white pawn"></span>
    <span class="piece e2 white pawn"></span>
    <span class="piece f2 white pawn"></span>
    <span class="piece g2 white pawn"></span>
    <span class="piece h2 white pawn"></span>
    <span class="piece a7 black pawn"></span>
    <span class="piece b7 black pawn"></span>
    <span class="piece c7 black pawn"></span>
    <span class="piece d7 black pawn"></span>
    <span class="piece e7 black pawn"></span>
    <span class="piece f7 black pawn"></span>
    <span class="piece g7 black pawn"></span>
    <span class="piece h7 black pawn"></span>
    <span class="piece a8 black rook"></span>
    <span class="piece b8 black knight"></span>
    <span class="piece c8 black bishop"></span>
    <span class="piece d8 black queen"></span>
    <span class="piece e8 black king"></span>
    <span class="piece f8 black bishop"></span>
    <span class="piece g8 black knight"></span>
    <span class="piece h8 black rook"></span>
</span>

<hr class="clear" />

`
}
