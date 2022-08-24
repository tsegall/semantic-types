# Semantic Type Reference Set #

The file 'reference.csv' captures the 'truth' as it relates to the Semantic Types from the data files stored in the data directory.

## Reference Fields  ##

The Reference file has a set of attributes keyed by the name of the file (File) and the field offset (FieldOffset)
 * File - the file processed
 * FieldOffset - the offset of the field within the record (0 origin)
 * Locale - the Locale of the file being processed
 * RecordCount - the number of records in the file
 * BaseType - e.g. Date, Long, String, ... 
 * SemanticType - the identified Semantic Type if determined
 * Notes - any notes to indicate observations made, typically to indicate why the field does not correspond to a Semantic Type

## F1-Score ##

Given a sample file 'current.csv' with the same layout as 'reference.csv' which is the output of any automated process that attempts to do automatic Semantic Type detection then the program performance will generate Precision, Recall, and F1-Scores by Semantic Type.

## Semantic Type Classification ##

The reference file which purports to be the arbiter of truth certainly has errors.  Some of these will be simple errors where fields have been incorrectly classified (if you see any feel free to raise an issue or better still a Pull Request), others may be the result of a close call.
