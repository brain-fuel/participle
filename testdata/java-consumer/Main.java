import dev.goforge.goplus.runtime.GpSlice;
import dev.goforge.goplus.runtime.GpString;
import dev.goforge.participle.Assignment;
import dev.goforge.participle.GpPackage;
import dev.goforge.participle.Outcome;
import dev.goforge.participle.Parsed;

public final class Main {
    public static void main(String[] args) {
        var grammar = GpPackage.AssignmentGrammar(42);
        var parser = GpPackage.BuildAssignments(grammar, GpPackage.AssignmentFirst(grammar));
        Outcome<GpSlice<Assignment>> outcome = GpPackage.Parse(parser, GpString.fromJava("answer=42"));
        if (!(outcome instanceof Parsed<GpSlice<Assignment>> parsed)) {
            throw new AssertionError("valid assignment was rejected");
        }
        Assignment assignment = parsed.Value.get(0);
        if (!assignment.Name.toJava().equals("answer") || assignment.Value != 42) {
            throw new AssertionError("wrong parsed assignment");
        }
        System.out.println("participle-java-consumer: ok");
    }
}
