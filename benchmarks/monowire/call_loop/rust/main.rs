use std::env;
use std::time::Instant;

#[derive(Clone, Copy)]
enum Op { Load, Move, Add, Sub, Jump, JumpZero, Call, Ret, Halt }
#[derive(Clone, Copy)]
struct Instr { op: Op, a: usize, b: usize, c: usize, imm: i64 }
struct Function { code: Vec<Instr>, regs: usize, args: usize }
#[derive(Clone, Copy)]
struct Frame { func: usize, base: usize, pc: usize, dest: usize }
struct Vm { functions: Vec<Function>, slots: Vec<i64>, frames: Vec<Frame> }

impl Vm {
    fn run(&mut self) -> i64 {
        self.frames.clear();
        self.frames.push(Frame { func: 0, base: 0, pc: 0, dest: 0 });
        self.slots.fill(0);
        loop {
            let frame = *self.frames.last().expect("empty frame stack");
            let function = &self.functions[frame.func];
            let ins = *function.code.get(frame.pc).expect("invalid pc");
            self.frames.last_mut().unwrap().pc += 1;
            let slots_len = self.slots.len();
            let at = |r: usize| -> usize {
                assert!(r < function.regs, "invalid register");
                let index = frame.base.checked_add(r).expect("register overflow");
                assert!(index < slots_len, "register capacity exceeded");
                index
            };
            match ins.op {
                Op::Load => self.slots[at(ins.a)] = ins.imm,
                Op::Move => self.slots[at(ins.a)] = self.slots[at(ins.b)],
                Op::Add => self.slots[at(ins.a)] = self.slots[at(ins.b)].wrapping_add(self.slots[at(ins.c)]),
                Op::Sub => self.slots[at(ins.a)] = self.slots[at(ins.b)].wrapping_sub(self.slots[at(ins.c)]),
                Op::Jump => self.frames.last_mut().unwrap().pc = ins.a,
                Op::JumpZero => if self.slots[at(ins.a)] == 0 { self.frames.last_mut().unwrap().pc = ins.b; },
                Op::Call => {
                    let callee = self.functions.get(ins.a).expect("invalid callee");
                    assert!(ins.b <= function.regs && callee.args <= function.regs - ins.b && ins.c < function.regs, "invalid call");
                    let base = frame.base.checked_add(ins.b).expect("base overflow");
                    assert!(base.checked_add(callee.regs).is_some_and(|end| end <= self.slots.len()), "register capacity exceeded");
                    self.frames.push(Frame { func: ins.a, base, pc: 0, dest: at(ins.c) });
                }
                Op::Ret => {
                    let result = self.slots[at(ins.a)];
                    let popped = self.frames.pop().unwrap();
                    if self.frames.is_empty() { return result; }
                    self.slots[popped.dest] = result;
                }
                Op::Halt => return self.slots[at(ins.a)],
            }
        }
    }
}
fn instr(op: Op, a: usize, b: usize, c: usize, imm: i64) -> Instr { Instr { op, a, b, c, imm } }
fn program(n: i64) -> Vec<Function> {
    use Op::*;
    vec![
        Function { regs: 7, args: 0, code: vec![
            instr(Load,0,0,0,0),instr(Load,1,0,0,n),instr(Load,2,0,0,1),
            instr(JumpZero,1,10,0,0),instr(Move,4,0,0,0),instr(Move,5,1,0,0),
            instr(Call,1,4,3,0),instr(Move,0,3,0,0),instr(Sub,1,1,2,0),
            instr(Jump,3,0,0,0),instr(Halt,0,0,0,0),
        ] },
        Function { regs: 3, args: 2, code: vec![instr(Add,2,0,1,0),instr(Ret,2,0,0,0)] },
    ]
}
fn main() {
    let args: Vec<String> = env::args().collect();
    let n: i64 = args.get(1).map(|s| s.parse().expect("invalid n")).unwrap_or(1000);
    let reps: usize = args.get(2).map(|s| s.parse().expect("invalid reps")).unwrap_or(20);
    assert!((0..=1_000_000).contains(&n) && reps > 0);
    let mut vm = Vm { functions: program(n), slots: vec![0; 16], frames: Vec::with_capacity(4) };
    let expected = n * (n+1) / 2;
    assert_eq!(vm.run(), expected);
    let start = Instant::now();
    let mut checksum = 0;
    for _ in 0..reps { checksum ^= vm.run(); }
    println!("language=rust workload=call_loop n={} reps={} elapsed_ns={} checksum={} expected={}", n, reps, start.elapsed().as_nanos(), checksum, expected);
}
