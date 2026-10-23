const std = @import("std");

// Linux x86_64 benchmark timer. Both variants use the same clock.
const Timespec = extern struct { tv_sec: isize, tv_nsec: isize };
extern "c" fn clock_gettime(clock_id: c_int, ts: *Timespec) c_int;
fn nowNs() !u64 {
    var ts: Timespec = undefined;
    if (clock_gettime(1, &ts) != 0) return error.ClockFailure; // CLOCK_MONOTONIC
    return @as(u64, @intCast(ts.tv_sec)) * 1_000_000_000 + @as(u64, @intCast(ts.tv_nsec));
}

const Op = enum { fetch, load, move, add, sub, jump, jump_zero, call, ret, halt };
const Instr = struct { op: Op, a: usize = 0, b: usize = 0, c: usize = 0, imm: i64 = 0 };
const Function = struct { code: []const Instr, regs: usize, args: usize };
const Frame = struct { func: usize, base: usize, pc: usize, dest: usize };

const Vm = struct {
    funcs: []const Function,
    slots: [20_016]i64 = @splat(0),
    frames: [10_016]Frame = undefined,
    depth: usize = 0,

    fn at(self: *Vm, f: Frame, r: usize) usize {
        std.debug.assert(r < self.funcs[f.func].regs);
        const index = f.base + r;
        std.debug.assert(index < self.slots.len);
        return index;
    }

    fn run(self: *Vm) i64 {
        self.slots = @splat(0);
        self.depth = 1;
        self.frames[0] = .{ .func = 0, .base = 0, .pc = 0, .dest = 0 };
        var f: Frame = undefined;
        var func: Function = undefined;
        var ins: Instr = undefined;
        dispatch: switch (Op.fetch) {
            .fetch => {
                f = self.frames[self.depth - 1];
                func = self.funcs[f.func];
                std.debug.assert(f.pc < func.code.len);
                ins = func.code[f.pc];
                self.frames[self.depth - 1].pc += 1;
                continue :dispatch ins.op;
            },
            .load => { self.slots[self.at(f, ins.a)] = ins.imm; continue :dispatch .fetch; },
            .move => { self.slots[self.at(f, ins.a)] = self.slots[self.at(f, ins.b)]; continue :dispatch .fetch; },
            .add => { self.slots[self.at(f, ins.a)] = self.slots[self.at(f, ins.b)] +% self.slots[self.at(f, ins.c)]; continue :dispatch .fetch; },
            .sub => { self.slots[self.at(f, ins.a)] = self.slots[self.at(f, ins.b)] -% self.slots[self.at(f, ins.c)]; continue :dispatch .fetch; },
            .jump => { self.frames[self.depth - 1].pc = ins.a; continue :dispatch .fetch; },
            .jump_zero => { if (self.slots[self.at(f, ins.a)] == 0) self.frames[self.depth - 1].pc = ins.b; continue :dispatch .fetch; },
            .call => {
                std.debug.assert(ins.a < self.funcs.len);
                const callee = self.funcs[ins.a];
                std.debug.assert(ins.b <= func.regs and callee.args <= func.regs - ins.b and ins.c < func.regs);
                const base = f.base + ins.b;
                std.debug.assert(base + callee.regs <= self.slots.len and self.depth < self.frames.len);
                self.frames[self.depth] = .{ .func = ins.a, .base = base, .pc = 0, .dest = self.at(f, ins.c) };
                self.depth += 1;
                continue :dispatch .fetch;
            },
            .ret => {
                const result = self.slots[self.at(f, ins.a)];
                self.depth -= 1;
                if (self.depth == 0) return result;
                self.slots[f.dest] = result;
                continue :dispatch .fetch;
            },
            .halt => return self.slots[self.at(f, ins.a)],
        }
    }
};

pub fn main(init: std.process.Init) !void {
    var args = std.process.Args.Iterator.init(init.minimal.args);
    _ = args.next();
    const n: i64 = if (args.next()) |arg| try std.fmt.parseInt(i64, arg, 10) else 1000;
    const reps: usize = if (args.next()) |arg| try std.fmt.parseInt(usize, arg, 10) else 20;
    if (n < 0 or n > 10_000 or reps == 0) return error.InvalidArguments;

    const caller = [_]Instr{.{ .op = .load, .a = 0, .b = 0, .c = 0, .imm = n },
.{ .op = .call, .a = 1, .b = 0, .c = 1, .imm = 0 },
.{ .op = .halt, .a = 1, .b = 0, .c = 0, .imm = 0 },};
    const callee = [_]Instr{.{ .op = .load, .a = 1, .b = 0, .c = 0, .imm = 1 },
.{ .op = .jump_zero, .a = 0, .b = 6, .c = 0, .imm = 0 },
.{ .op = .sub, .a = 2, .b = 0, .c = 1, .imm = 0 },
.{ .op = .call, .a = 1, .b = 2, .c = 1, .imm = 0 },
.{ .op = .add, .a = 1, .b = 0, .c = 1, .imm = 0 },
.{ .op = .ret, .a = 1, .b = 0, .c = 0, .imm = 0 },
.{ .op = .load, .a = 1, .b = 0, .c = 0, .imm = 0 },
.{ .op = .ret, .a = 1, .b = 0, .c = 0, .imm = 0 },};
    const funcs = [_]Function{
        .{ .code = &caller, .regs = 2, .args = 0 },
        .{ .code = &callee, .regs = 3, .args = 1 },
    };
    var vm = Vm{ .funcs = &funcs };
    const expected = @divTrunc(n * (n + 1), 2);
    if (vm.run() != expected) return error.WrongResult;
    const start = try nowNs();
    var checksum: i64 = 0;
    for (0..reps) |_| checksum ^= vm.run();
    const elapsed = (try nowNs()) - start;
    const expected_checksum: i64 = if (reps % 2 == 0) 0 else expected;
    if (checksum != expected_checksum) return error.WrongChecksum;
    std.debug.print("language=zig variant=labeled workload=recursion n={d} reps={d} elapsed_ns={d} checksum={d} expected={d}\n", .{ n, reps, elapsed, checksum, expected });
}
