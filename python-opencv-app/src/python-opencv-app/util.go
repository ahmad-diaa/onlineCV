package main

import(
  "strings"
)
func contains(target string, filenames []string) bool{
  for _, value := range filenames {
      if value == target {
        return true
      }
  }
  return false
}

func isImage(filename string) bool {
  if strings.HasSuffix(filename, "jpg") {
    return true
  }
  if strings.HasSuffix(filename, "png") {
    return true
  }
  if strings.HasSuffix(filename, "jpeg") {
      return true
  }
  return false
}
