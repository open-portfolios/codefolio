# Facade

Provide a unified interface to a set of interfaces in a subsystem. Facade defines a higher-level interface that makes the subsystem easier to use.

## Intent

- Provide a unified interface to a set of interfaces in a subsystem
- Make a subsystem easier to use
- Wrap a complex subsystem with a simpler interface

## Implementation

```java
public class CPU {
    public void freeze() {
        System.out.println("CPU freeze");
    }
    
    public void jump(long position) {
        System.out.println("CPU jump to " + position);
    }
    
    public void execute() {
        System.out.println("CPU execute");
    }
}

public class Memory {
    public void load(long position, byte[] data) {
        System.out.println("Memory load at " + position);
    }
}

public class HardDrive {
    public byte[] read(long lba, int size) {
        System.out.println("HardDrive read " + size + " bytes");
        return new byte[size];
    }
}

// Facade
public class ComputerFacade {
    private CPU cpu;
    private Memory memory;
    private HardDrive hardDrive;
    
    public ComputerFacade() {
        this.cpu = new CPU();
        this.memory = new Memory();
        this.hardDrive = new HardDrive();
    }
    
    public void start() {
        cpu.freeze();
        memory.load(0, hardDrive.read(0, 1024));
        cpu.jump(0);
        cpu.execute();
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        ComputerFacade computer = new ComputerFacade();
        computer.start();
    }
}
```

## When to Use

- When you want to provide a simple interface to a complex subsystem
- When there are many dependencies between clients and implementation classes
- When you want to layer your subsystems
