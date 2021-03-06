package chess

import (
	"fmt"
	"strings"
)

// Encoder is the interface implemented by objects that can
// encode a move into a string given the position.  It is not
// the encoders responsibility to validate the move.
type Encoder interface {
	Encode(pos *Position, m *Move) string
}

// Decoder is the interface implemented by objects that can
// decode a string into a move given the position. It is not
// the decoders responsibility to validate the move.  An error
// is returned if the string could not be decoded.
type Decoder interface {
	Decode(pos *Position, s string) (*Move, error)
}

// Notation is the interface implemented by objects that can
// encode and decode moves.
type Notation interface {
	Encoder
	Decoder
}

// UCINotation is a more computer friendly alternative to algebraic
// notation.  This notation uses the same format as the UCI (Universal Chess
// Interface).  Examples: e2e4, e7e5, e1g1 (white short castling), e7e8q (for promotion)
type UCINotation struct{}

// String implements the fmt.Stringer interface and returns
// the notation's name.
func (UCINotation) String() string {
	return "UCI Notation"
}

// Encode implements the Encoder interface.
func (UCINotation) Encode(pos *Position, m *Move) string {
	return m.GetS1().String() + m.GetS2().String() + m.Promo().String()
}

// Decode implements the Decoder interface.
func (UCINotation) Decode(pos *Position, s string) (*Move, error) {
	l := len(s)
	err := fmt.Errorf(`chess: failed to decode long algebraic notation text "%s" for position %s`, s, pos)
	if l < 4 || l > 5 {
		return nil, err
	}
	S1, ok := strToSquareMap[s[0:2]]
	if !ok {
		return nil, err
	}
	S2, ok := strToSquareMap[s[2:4]]
	if !ok {
		return nil, err
	}
	promo := NoPieceType
	if l == 5 {
		promo = pieceTypeFromChar(s[4:5])
		if promo == NoPieceType {
			return nil, err
		}
	}
	m := &Move{S1: S1, S2: S2, promo: promo}
	if pos == nil {
		return m, nil
	}
	p := pos.Board().Piece(S1)
	if p.Type() == King {
		if (S1 == E1 && S2 == G1) || (S1 == E8 && S2 == G8) {
			m.addTag(KingSideCastle)
		} else if (S1 == E1 && S2 == C1) || (S1 == E8 && S2 == C8) {
			m.addTag(QueenSideCastle)
		}
	} else if p.Type() == Pawn && S2 == pos.enPassantSquare {
		m.addTag(EnPassant)
		m.addTag(Capture)
	}
	c1 := p.Color()
	c2 := pos.Board().Piece(S2).Color()
	if c2 != NoColor && c1 != c2 {
		m.addTag(Capture)
	}
	return m, nil
}

// AlgebraicNotation (or Standard Algebraic Notation) is the
// official chess notation used by FIDE. Examples: e4, e5,
// O-O (short castling), e8=Q (promotion)
type AlgebraicNotation struct{}

// String implements the fmt.Stringer interface and returns
// the notation's name.
func (AlgebraicNotation) String() string {
	return "Algebraic Notation"
}

// Encode implements the Encoder interface.
func (AlgebraicNotation) Encode(pos *Position, m *Move) string {
	checkChar := getCheckChar(pos, m)
	if m.HasTag(KingSideCastle) {
		return "O-O" + checkChar
	} else if m.HasTag(QueenSideCastle) {
		return "O-O-O" + checkChar
	}
	p := pos.Board().Piece(m.GetS1())
	pChar := charFromPieceType(p.Type())
	S1Str := formS1(pos, m)
	capChar := ""
	if m.HasTag(Capture) || m.HasTag(EnPassant) {
		capChar = "x"
		if p.Type() == Pawn && S1Str == "" {
			capChar = m.S1.File().String() + "x"
		}
	}
	promoText := charForPromo(m.promo)
	return pChar + S1Str + capChar + m.S2.String() + promoText + checkChar
}

// Decode implements the Decoder interface.
func (AlgebraicNotation) Decode(pos *Position, s string) (*Move, error) {
	s = removeSubstrings(s, "?", "!", "+", "#", "e.p.")
	for _, m := range pos.ValidMoves() {
		str := AlgebraicNotation{}.Encode(pos, m)
		str = removeSubstrings(str, "?", "!", "+", "#", "e.p.")
		if str == s {
			return m, nil
		}
	}
	return nil, fmt.Errorf("chess: could not decode algebraic notation %s for position %s", s, pos.String())
}

// LongAlgebraicNotation is a fully expanded version of
// algebraic notation in which the starting and ending
// squares are specified.
// Examples: e2e4, Rd3xd7, O-O (short castling), e7e8=Q (promotion)
type LongAlgebraicNotation struct{}

// String implements the fmt.Stringer interface and returns
// the notation's name.
func (LongAlgebraicNotation) String() string {
	return "Long Algebraic Notation"
}

// Encode implements the Encoder interface.
func (LongAlgebraicNotation) Encode(pos *Position, m *Move) string {
	checkChar := getCheckChar(pos, m)
	if m.HasTag(KingSideCastle) {
		return "O-O" + checkChar
	} else if m.HasTag(QueenSideCastle) {
		return "O-O-O" + checkChar
	}
	p := pos.Board().Piece(m.GetS1())
	pChar := charFromPieceType(p.Type())
	S1Str := m.S1.String()
	capChar := ""
	if m.HasTag(Capture) || m.HasTag(EnPassant) {
		capChar = "x"
		if p.Type() == Pawn && S1Str == "" {
			capChar = m.S1.File().String() + "x"
		}
	}
	promoText := charForPromo(m.promo)
	return pChar + S1Str + capChar + m.S2.String() + promoText + checkChar
}

// Decode implements the Decoder interface.
func (LongAlgebraicNotation) Decode(pos *Position, s string) (*Move, error) {
	s = removeSubstrings(s, "?", "!", "+", "#", "e.p.")
	for _, m := range pos.ValidMoves() {
		str := LongAlgebraicNotation{}.Encode(pos, m)
		str = removeSubstrings(str, "?", "!", "+", "#", "e.p.")
		if str == s {
			return m, nil
		}
	}
	return nil, fmt.Errorf("chess: could not decode long algebraic notation %s for position %s", s, pos.String())
}

func getCheckChar(pos *Position, move *Move) string {
	if !move.HasTag(Check) {
		return ""
	}
	nextPos := pos.Update(move)
	if nextPos.Status() == Checkmate {
		return "#"
	}
	return "+"
}

func formS1(pos *Position, m *Move) string {
	p := pos.board.Piece(m.S1)
	if p.Type() == Pawn {
		return ""
	}

	var req, fileReq, rankReq bool
	moves := pos.ValidMoves()

	for _, mv := range moves {
		if mv.S1 != m.S1 && mv.S2 == m.S2 && p == pos.board.Piece(mv.S1) {
			req = true

			if mv.S1.File() == m.S1.File() {
				rankReq = true
			}

			if mv.S1.Rank() == m.S1.Rank() {
				fileReq = true
			}
		}
	}

	var S1 = ""

	if fileReq || !rankReq && req {
		S1 = m.S1.File().String()
	}

	if rankReq {
		S1 += m.S1.Rank().String()
	}

	return S1
}

func charForPromo(p PieceType) string {
	c := charFromPieceType(p)
	if c != "" {
		c = "=" + c
	}
	return c
}

func charFromPieceType(p PieceType) string {
	switch p {
	case King:
		return "K"
	case Queen:
		return "Q"
	case Rook:
		return "R"
	case Bishop:
		return "B"
	case Knight:
		return "N"
	}
	return ""
}

func pieceTypeFromChar(c string) PieceType {
	switch c {
	case "q":
		return Queen
	case "r":
		return Rook
	case "b":
		return Bishop
	case "n":
		return Knight
	}
	return NoPieceType
}

func removeSubstrings(s string, subs ...string) string {
	for _, sub := range subs {
		s = strings.Replace(s, sub, "", -1)
	}
	return s
}
