package main

import (
 "fmt"
 "os"
 "strconv"
 "time"
)

type Op uint8
const ( Load Op = iota; Move; Add; Sub; Jump; JumpZero; Call; Ret; Halt )
type Instr struct { op Op; a,b,c int; imm int64 }
type Function struct { code []Instr; regs, args int }
type Frame struct { fn, base, pc, dest int }
type VM struct { funcs []Function; slots []int64; frames []Frame }
func (v *VM) run() int64 {
 v.frames = append(v.frames[:0], Frame{fn:0,base:0,pc:0,dest:-1})
 for i:=range v.slots { v.slots[i]=0 }
 for len(v.frames)>0 {
  f:=&v.frames[len(v.frames)-1]; fn:=v.funcs[f.fn]
  if f.pc<0 || f.pc>=len(fn.code) { panic("invalid pc") }
  ins:=fn.code[f.pc]; f.pc++
  at:=func(r int) int { if r<0 || r>=fn.regs {panic("invalid register")}; return f.base+r }
  switch ins.op {
  case Load: v.slots[at(ins.a)]=ins.imm
  case Move: v.slots[at(ins.a)]=v.slots[at(ins.b)]
  case Add: v.slots[at(ins.a)]=v.slots[at(ins.b)]+v.slots[at(ins.c)]
  case Sub: v.slots[at(ins.a)]=v.slots[at(ins.b)]-v.slots[at(ins.c)]
  case Jump: f.pc=ins.a
  case JumpZero: if v.slots[at(ins.a)]==0 { f.pc=ins.b }
  case Call:
   target:=ins.a; start:=ins.b; dest:=ins.c
   if target<0||target>=len(v.funcs) {panic("invalid callee")}
   callee:=v.funcs[target]
   if start<0||start+callee.args>fn.regs || dest<0||dest>=fn.regs {panic("invalid call")}
   base:=f.base+start
   // The callee may extend beyond the caller's register window.
   if base+callee.regs>len(v.slots) {panic("register capacity exceeded")}
   v.frames=append(v.frames,Frame{fn:target,base:base,pc:0,dest:f.base+dest})
  case Ret:
   result:=v.slots[at(ins.a)]; dest:=f.dest
   v.frames=v.frames[:len(v.frames)-1]
   if len(v.frames)==0 {return result}
   v.slots[dest]=result
  case Halt: return v.slots[at(ins.a)]
  default: panic("invalid opcode")
  }
 }
 panic("unreachable")
}
func program(n int64) []Function { return []Function{
{regs:2,args:0,code:[]Instr{{op:Load,a:0,b:0,c:0,imm:n},
{op:Call,a:1,b:0,c:1,imm:0},
{op:Halt,a:1,b:0,c:0,imm:0},}},
{regs:3,args:1,code:[]Instr{{op:Load,a:1,b:0,c:0,imm:1},
{op:JumpZero,a:0,b:6,c:0,imm:0},
{op:Sub,a:2,b:0,c:1,imm:0},
{op:Call,a:1,b:2,c:1,imm:0},
{op:Add,a:1,b:0,c:1,imm:0},
{op:Ret,a:1,b:0,c:0,imm:0},
{op:Load,a:1,b:0,c:0,imm:0},
{op:Ret,a:1,b:0,c:0,imm:0},}},
} }
func main(){
 n:=int64(1000); reps:=20
 if len(os.Args)>1 {x,e:=strconv.ParseInt(os.Args[1],10,64);if e!=nil||x<0||x>10000{panic("bad n")};n=x}
 if len(os.Args)>2 {x,e:=strconv.Atoi(os.Args[2]);if e!=nil||x<1{panic("bad reps")};reps=x}
 v:=VM{funcs:program(n),slots:make([]int64,2*int(n)+8),frames:make([]Frame,0,int(n)+2)}
 expected:=n*(n+1)/2
 if got:=v.run();got!=expected {panic(fmt.Sprintf("incorrect result %d != %d",got,expected))}
 start:=time.Now();var check int64
 for i:=0;i<reps;i++ {check^=v.run()}
 fmt.Printf("language=go workload=recursion n=%d reps=%d elapsed_ns=%d checksum=%d expected=%d\n",n,reps,time.Since(start).Nanoseconds(),check,expected)
}
