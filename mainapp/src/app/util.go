package main

func sliceContains(target string, filenames []string) bool{
  for _, value := range filenames {
      if value == target {
        return true
      }
  }
  return false
}
