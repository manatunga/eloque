const std = @import("std");

const Op = enum { load, move, add, sub, jump, jump_zero, call, ret, halt };

const Instr = struct {
    op: Op,
    a: usize = 0,
    b: usize = 0,
    c: usize = 0,
    imm: i64 = 0,
};

const Function = struct {
    code: []const Instr,
    regs: usize,
    args: usize,
};

const Frame = struct {
    func: usize,
    base: usize,
    pc: usize,
    dest: usize,
};

const Vm = struct {
    funcs: []const Function,
    slots: [16]i64 = @as([16]i64, @splat(0)),
    frames: [32]Frame = undefined,
    depth: usize = 0,

    fn at(self: *Vm, f: Frame, r: usize) usize {
        std.debug.assert(r < self.funcs[f.func].regs);
        const index = f.base + r;
        std.debug.assert(index < self.slots.len);
        return index;
    }

    fn run(self: *Vm) i64 {
        self.slots = @as([16]i64, @splat(0));
        self.depth = 1;
        self.frames[0] = .{
            .func = 0,
            .base = 0,
            .pc = 0,
            .dest = 0,
        };

        while (true) {
            const f = self.frames[self.depth - 1];
            const func = self.funcs[f.func];

            std.debug.assert(f.pc < func.code.len);
            const ins = func.code[f.pc];
            self.frames[self.depth - 1].pc += 1;

            switch (ins.op) {
                .load => self.slots[self.at(f, ins.a)] = ins.imm,
                .move => self.slots[self.at(f, ins.a)] = self.slots[self.at(f, ins.b)],
                .add => self.slots[self.at(f, ins.a)] =
                    self.slots[self.at(f, ins.b)] +%
                    self.slots[self.at(f, ins.c)],
                .sub => self.slots[self.at(f, ins.a)] =
                    self.slots[self.at(f, ins.b)] -%
                    self.slots[self.at(f, ins.c)],
                .jump => self.frames[self.depth - 1].pc = ins.a,
                .jump_zero => {
                    if (self.slots[self.at(f, ins.a)] == 0) {
                        self.frames[self.depth - 1].pc = ins.b;
                    }
                },
                .call => {
                    std.debug.assert(ins.a < self.funcs.len);
                    const callee = self.funcs[ins.a];

                    std.debug.assert(
                        ins.b <= func.regs and
                            callee.args <= func.regs - ins.b and
                            ins.c < func.regs,
                    );

                    const base = f.base + ins.b;

                    std.debug.assert(
                        base + callee.regs <= self.slots.len and
                            self.depth < self.frames.len,
                    );

                    self.frames[self.depth] = .{
                        .func = ins.a,
                        .base = base,
                        .pc = 0,
                        .dest = self.at(f, ins.c),
                    };
                    self.depth += 1;
                },
                .ret => {
                    const result = self.slots[self.at(f, ins.a)];
                    self.depth -= 1;

                    if (self.depth == 0) return result;
                    self.slots[f.dest] = result;
                },
                .halt => return self.slots[self.at(f, ins.a)],
            }
        }
    }
};

pub fn main(init: std.process.Init) !void {
    var args = std.process.Args.Iterator.init(init.minimal.args);
    _ = args.next(); // Skip executable name.

    const n: i64 = if (args.next()) |arg|
        try std.fmt.parseInt(i64, arg, 10)
    else
        1000;

    const reps: usize = if (args.next()) |arg|
        try std.fmt.parseInt(usize, arg, 10)
    else
        20;

    if (n < 0 or n > 1_000_000 or reps == 0) {
        return error.InvalidArguments;
    }

    const caller = [_]Instr{
        .{ .op = .load, .a = 0, .imm = 0 },
        .{ .op = .load, .a = 1, .imm = n },
        .{ .op = .load, .a = 2, .imm = 1 },
        .{ .op = .jump_zero, .a = 1, .b = 10 },
        .{ .op = .move, .a = 4, .b = 0 },
        .{ .op = .move, .a = 5, .b = 1 },
        .{ .op = .call, .a = 1, .b = 4, .c = 3 },
        .{ .op = .move, .a = 0, .b = 3 },
        .{ .op = .sub, .a = 1, .b = 1, .c = 2 },
        .{ .op = .jump, .a = 3 },
        .{ .op = .halt, .a = 0 },
    };

    const callee = [_]Instr{
        .{ .op = .add, .a = 2, .b = 0, .c = 1 },
        .{ .op = .ret, .a = 2 },
    };

    const funcs = [_]Function{
        .{ .code = &caller, .regs = 7, .args = 0 },
        .{ .code = &callee, .regs = 3, .args = 2 },
    };

    var vm = Vm{ .funcs = &funcs };

    const expected = @divTrunc(n * (n + 1), 2);

    if (vm.run() != expected) {
        return error.WrongResult;
    }

    var checksum: i64 = 0;
    for (0..reps) |_| {
        checksum ^= vm.run();
    }

    const expected_checksum: i64 =
        if (reps % 2 == 0) 0 else expected;

    if (checksum != expected_checksum) {
        return error.WrongChecksum;
    }

    std.debug.print(
        "language=zig workload=call_loop n={d} reps={d} checksum={d} expected={d}\n",
        .{ n, reps, checksum, expected },
    );
}
