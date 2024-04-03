import { Component, ElementRef, EventEmitter, Input, OnInit, Output, Renderer2, ViewChild } from '@angular/core';
import { UntypedFormControl } from '@angular/forms';
import { timer } from 'rxjs';

@Component({
  selector: 'app-searchbox',
  templateUrl: './searchbox.component.html',
  styleUrls: ['./searchbox.component.scss']
})
export class SearchboxComponent implements OnInit {

  @ViewChild('container', { static: true }) container: any;
  @ViewChild('search', { static: true }) search: any;

  @Input() placeholder: string;
  @Input() color: string;
  @Input() icon: string;
  @Input() ourValue: string;

  @Output() onAction = new EventEmitter<string>();
  @Output() valueChange = new EventEmitter<string>();
  @Output() onSubmit = new EventEmitter<string>();

  public searchControl = new UntypedFormControl('');

  public hasFocus: boolean;
  public hasValue: boolean;
  public searchDisabled: boolean;

  constructor(
    private renderer: Renderer2,
    private searchBox: ElementRef
  ) { }

  ngOnInit() {
    this.lower();
    if (this.ourValue && this.ourValue.length > 0) {
      this.searchControl.setValue(this.ourValue);
      this.searchControl.markAsDirty();
      this.search.nativeElement.focus();
      this.hasValue = true;
      this.raise();
    }
    this.searchDisabled = this.onSubmit.observers.length === 0;
    this.searchControl.valueChanges.subscribe((value: any) => {
      this.hasValue = value.length > 0;
      this.ourValue = value;
      this.valueChange.emit(value);
    });
    // Capture Mouse Over Events
    this.renderer.listen(this.searchBox.nativeElement, 'mouseover', () => this.raise());
    this.renderer.listen(this.searchBox.nativeElement, 'mouseleave', () => this.lower());
    // Capture Input Focus Events
    this.renderer.listen(this.search.nativeElement, 'focus', () => this.hasFocus = true);
    this.renderer.listen(this.search.nativeElement, 'blur', () => {
      this.hasFocus = false;
      timer(500).subscribe(() => { this.lower(); });
    });
  }

  @Input()
  get value() {
    return this.ourValue;
  }

  set value(val) {
    this.searchControl.setValue(val);
  }

  submit() {
    this.onSubmit.emit(this.searchControl.value); }

  clear() { this.searchControl.setValue(''); }

  raise() {
    this.renderer.addClass(this.container.nativeElement, 'mat-elevation-z4');
    this.renderer.removeClass(this.container.nativeElement, 'mat-elevation-z1');
  }

  lower() {
    if (!this.hasFocus) {
      this.renderer.removeClass(this.container.nativeElement, 'mat-elevation-z4');
      this.renderer.addClass(this.container.nativeElement, 'mat-elevation-z1');
    }
  }
}
